package orm

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Kotodian/gokit/business"
	"github.com/Kotodian/gokit/datasource/rabbitmq"
	"github.com/streadway/amqp"
	"google.golang.org/grpc/metadata"

	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/gokit/datasource/grpc"
	"github.com/Kotodian/gokit/sync/errgroup.v2"
	proto "github.com/Kotodian/protocol/golang/xa"

	"github.com/edwardhey/gorm"
	"github.com/sirupsen/logrus"
	libgrpc "google.golang.org/grpc"

	"github.com/Kotodian/gokit/datasource/redis"
)

var xas sync.Map
var Exchange rabbitmq.Exchange

func init() {
	Exchange = business.XAExchange
}

type XA struct {
	ID       datasource.UUID
	BranchID datasource.UUID
	DB       *gorm.DB
	log      *logrus.Entry
	//Commit bool
}

type XAListen struct {
	Commit bool
	Err    error
}

func (xa XA) GetXID() string {
	return fmt.Sprintf("%s:%s:goiot", xa.ID, xa.BranchID)
}

//func NewXACaller(ctx context.Context, id string) *XA {
//	xa := &XA{
//		ID:     id,
//		Module: db.Dialect().CurrentDatabase(),
//		DB:     db.New(),
//	}
//	xa.DB.InstantSet("caller:xa", true)
//	//ctx = context.WithValue(ctx, "xa", xa)
//	return xa
//}

func NewMasterXA(id datasource.UUID, level ...sql.IsolationLevel) (context.Context, *XA, error) {
	branchID := <-UUID
	xa := &XA{
		ID: id,
		//Module: db.Dialect().CurrentDatabase(),
		BranchID: branchID,
		//DB:    db.New().Debug(),
		log: logrus.WithFields(logrus.Fields{
			"module": "xa",
			"gtid":   id,
			"bid":    branchID,
		}),
	}
	ctx := context.WithValue(context.Background(), "xid", xa.GetXID())

	if err := xa.Begin(ctx); err != nil {
		return nil, nil, err
	}
	txOptions := &sql.TxOptions{}
	if len(level) > 0 {
		txOptions.Isolation = level[0]
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("xid", id.String(), "isolation_level", fmt.Sprintf("%d", txOptions.Isolation)))
	} else {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("xid", id.String()))
	}
	xa.DB = db.BeginTx(ctx, txOptions)
	if xa.DB.Error != nil {
		return nil, nil, xa.DB.Error
	}
	ctx = context.WithValue(ctx, "db", xa.DB)
	//fmt.Println("xa master", xa.GetXID())
	//xa.DB.InstantSet("db", xa.DB)
	//xa, _ := xas.LoadOrStore(id, )
	//xas.Store(xa.GetXID(), true)
	return ctx, xa, nil
}

func NewSlaveXAWithContext(ctx context.Context) (context.Context, *XA, error) {
	//grpc ctx携带metadata
	var xid datasource.UUID
	md, _ := metadata.FromIncomingContext(ctx)
	if xidArr := md.Get("xid"); len(xidArr) > 0 {
		xid, _ = datasource.ParseUUID(xidArr[0])
	}
	if xid <= 0 {
		return ctx, nil, fmt.Errorf("xid不能为0")
	}

	if isolationLevels := md.Get("isolation_level"); len(isolationLevels) > 0 {
		isolationLevel, _ := datasource.ParseUUID(isolationLevels[0])
		return NewSlaveXA(xid, sql.IsolationLevel(isolationLevel))
	}
	return NewSlaveXA(xid)
}

func NewSlaveXA(id datasource.UUID, level ...sql.IsolationLevel) (context.Context, *XA, error) {
	branchID := <-UUID

	xa := &XA{
		ID: id,
		//Module: db.Dialect().CurrentDatabase(),
		BranchID: branchID,
		//DB:    db.New().Debug(),
		log: logrus.WithFields(logrus.Fields{
			"gtid": id,
			"bid":  branchID,
		}),
	}
	ctx := context.WithValue(context.Background(), "xid", xa.GetXID())

	txOptions := &sql.TxOptions{}
	if len(level) > 0 {
		txOptions.Isolation = level[0]
	}

	xa.DB = db.BeginTx(ctx, txOptions)
	if xa.DB.Error != nil {
		return nil, nil, xa.DB.Error
	}
	ctx = context.WithValue(ctx, "db", xa.DB)
	//xa.DB.InstantSet("db", xa.DB)
	//xa, _ := xas.LoadOrStore(id, )
	//xas.Store(xa.GetXID(), true)
	return ctx, xa, nil
}

func (xa *XA) ExecEnd() error {
	return xa.DB.Exec(fmt.Sprintf("xa end '%s'", xa.GetXID())).Error
}

func (xa *XA) ExecPrepare() error {
	return xa.DB.Exec(fmt.Sprintf("xa prepare '%s'", xa.GetXID())).Error
}

func (xa *XA) ExecCommit() error {
	//if err := xa.DB.Exec(fmt.Sprintf("xa commit '%s'", xa.GetXID())).Error; err != nil {
	//	return err
	//}
	return xa.DB.Commit().Error
}

func (xa *XA) Exec(sql string, values ...interface{}) error {
	return xa.DB.Exec(sql, values...).Error
}

func (xa *XA) ExecRollback() error {
	//if err := xa.DB.Exec(fmt.Sprintf("xa rollback '%s'", xa.GetXID())).Error; err != nil {
	//	return err
	//}
	return xa.DB.Rollback().Error
}

func (xa *XA) CommitOrRollback(ctx context.Context, e error) (err error) {
	var end bool
	commit := false
	defer func() {
		if err != nil {
			xa.log.Error(err)
			if end {
				if rollbackErr := xa.ExecRollback(); rollbackErr != nil {
					xa.log.Error("transaction rollback error,", rollbackErr)
				}
				commit = false
			}
		} else if err = xa.ExecCommit(); err != nil {
			xa.log.Error("transaction commit error,", err)
		} else {
			commit = true
		}

		b, _ := json.Marshal(commit)

		//发布给所有其他事物处理器，统一进行回滚或提交
		if e := Exchange.Publish(ctx, fmt.Sprintf("notify:%d", xa.ID), amqp.Publishing{
			MessageId: xa.ID.String(),
			Body:      b,
		}); e != nil {
			xa.log.Error("transaction notify error,", e)
		}
		//if _, e := redis.Publish(fmt.Sprintf("%s:goiot:xa", xa.ID), commit); e != nil {
		//	//if xaErr := xa.ExecRollback(); xaErr != nil {
		//	//	err = fmt.Errorf("%s,%s,%s", err, e, xaErr)
		//	xa.log.Error(err)
		//}
		/*else {
			err = xa.ExecCommit()
		}*/
		//redis.Publish()
	}()

	if err = xa.ExecEnd(); err != nil {
		return
	}
	end = true
	act := proto.ACT_COMMIT
	if e == nil {
		if err = xa.ExecPrepare(); err != nil {
			return
		}
	} else {
		act = proto.ACT_ROLLBACK
	}

	req := &proto.XAReq{
		ID:  xa.ID.Uint64(),
		Act: act,
	}
	//fmt.Println("xa rpc", req)
	resp := &proto.XAResp{}
	ctx, _ = context.WithTimeout(ctx, 3*time.Second)
	if err = grpc.Invoke(ctx, proto.RPCServicesServer.XA, req, resp); err != nil {
		err = fmt.Errorf(libgrpc.ErrorDesc(err))
		return
	} else {
		err = e
	}

	return
}

func (xa *XA) ListenCommitOrRollback(ctx context.Context, e error, commitCallbackFn ...func()) (xl chan XAListen, err error) {
	var end bool
	defer func() {
		if err != nil {
			xa.log.Error(err)
			if end {
				if rollbackErr := xa.ExecRollback(); rollbackErr != nil {
					xa.log.Error("transaction rollback error,", rollbackErr)
				}
			}
		}
	}()

	if err = xa.ExecEnd(); err != nil {
		return nil, err
	}
	end = true
	if e == nil {
		if err = xa.ExecPrepare(); err != nil {
			return nil, err
		}
		//fmt.Println("xa master commit")
		xl = make(chan XAListen, 1)
		go func(ctx context.Context) {
			var commit bool
			defer func() {
				//fmt.Println("xa slave defer 1 ", err)
				if err != nil {
					xl <- XAListen{
						Commit: commit,
						Err:    err,
					}
				}
				close(xl)
			}()

			for {
				g := errgroup.WithTimeout(ctx, 10*time.Second)
				g.Go(func(ctx context.Context) (err error) {
					var conn *amqp.Connection
					if conn, err = amqp.Dial(rabbitmq.MQURL); err != nil {
						return
					}
					defer func() {
						_ = conn.Close()
					}()
					var channel *amqp.Channel
					if channel, err = conn.Channel(); err != nil {
						return
					}
					defer func() {
						_ = channel.Close()
					}()
					ctx = context.WithValue(ctx, "conn", conn)
					ctx = context.WithValue(ctx, "channel", channel)

					//fmt.Println("------------------------ ListenCommitOrRollback")
					Exchange.Start(ctx, 1, rabbitmq.Receiver{
						Name:      fmt.Sprintf("xa_notify_%d_%d", xa.ID, <-UUID),
						RouterKey: fmt.Sprintf("notify:%d", xa.ID),
						Exclusive: true,
						QOSFn: func() (prefetchCount, prefetchSize int, global bool) {
							return 1, 0, true
						},
						OnReceiveFn: func(r rabbitmq.Receiver, msg amqp.Delivery) (ret bool) {
							defer func() {
								r.QuitFunc()
							}()
							ret = true
							logEntry := logrus.WithFields(logrus.Fields{
								"act": "xa listen",
								"xid": xa.ID,
							})
							logEntry.Info(string(msg.Body))
							//var commit bool
							if err = json.Unmarshal(msg.Body, &commit); err != nil {
								logEntry.Error(err)
								return
							}
							xl <- XAListen{
								Commit: commit,
								Err:    nil,
							}
							//退出监听
							//r.QuitFunc()
							//退出监听
							return
						},
					})
					return err
				})
				if err = g.Wait(); err != nil {
					xa.log.Error("sub xa error, ", err)
					//去xa服务器获取
					var act proto.ACT
					if act, err = xa.GetACT(ctx); err != nil {
						xa.log.Error("get xa act error, ", err)
						if err.Error() == "404" {
							return
						}
						continue
					} else if act == proto.ACT_PREPARE {
						xa.log.Warnf("xa prepare, wait for commit or rollback")
						continue
					}
					xa.log.Info("get xa act ", act)
					commit = act == proto.ACT_COMMIT
				}
				xa.log.Info("commit:", commit)
				if commit {
					if e := xa.ExecCommit(); e != nil {
						xa.log.Error("xa commit error, %s", e.Error())
						if e := xa.ExecRollback(); e != nil {
							xa.log.Error("xa rollback error, %s", e.Error())
							continue
							//return
						}
					} else if len(commitCallbackFn) > 0 {
						time.Sleep(time.Second)
						commitCallbackFn[0]()
					}
				} else if e := xa.ExecRollback(); e != nil {
					xa.log.Error("xa rollback error, %s", e.Error())
					continue
					//return
				}
				return
			}
		}(ctx)
	} else {
		//xa.ExecRollback()
		err = e
	}
	return
}

func (xa *XA) Begin(ctx context.Context) (err error) {
	req := &proto.XAReq{
		ID:  xa.ID.Uint64(),
		Act: proto.ACT_PREPARE,
	}
	resp := &proto.XAResp{}
	ctx, _ = context.WithTimeout(ctx, 3*time.Second)
	if err = grpc.Invoke(ctx, proto.RPCServicesServer.XA, req, resp); err != nil {
		err = fmt.Errorf(libgrpc.ErrorDesc(err))
		return
	}
	return
}

func (xa *XA) GetACT(ctx context.Context) (act proto.ACT, err error) {
	req := &proto.GetXAReq{
		ID: xa.ID.Uint64(),
	}
	resp := &proto.GetXAResp{}
	ctx, _ = context.WithTimeout(ctx, 3*time.Second)
	if err = grpc.Invoke(ctx, proto.RPCServicesServer.GetXA, req, resp); err != nil {
		err = fmt.Errorf(libgrpc.ErrorDesc(err))
		return
	}
	act = resp.Act
	return
}

func (xa *XA) Rollback(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			xa.log.Error(err)
		}
		_ = xa.ExecRollback()
		if _, err = redis.Publish(fmt.Sprintf("%s:goiot:xa", xa.ID), false); err != nil {
			return
		}
	}()
	//fmt.Println("xa master rollback")
	req := &proto.XAReq{
		ID:  xa.ID.Uint64(),
		Act: proto.ACT_ROLLBACK,
	}
	resp := &proto.XAResp{}
	ctx, _ = context.WithTimeout(ctx, 3*time.Second)
	if err = grpc.Invoke(ctx, proto.RPCServicesServer.XA, req, resp); err != nil {
		err = fmt.Errorf(libgrpc.ErrorDesc(err))
		return
	}
	return
}
