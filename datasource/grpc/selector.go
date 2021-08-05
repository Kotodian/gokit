package grpc

// import (
// 	"errors"
// 	"fmt"
// 	"strings"

// 	log "github.com/sirupsen/logrus"
// 	grpclb "github.com/edwardhey/grpc-lb"
// 	"github.com/lafikl/consistent"
// 	"golang.org/x/net/context"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/metadata"
// )

// type Selector struct {
// 	grpclb.BaseSelector
// 	hash     *consistent.Consistent
// 	logEntry *log.Entry
// 	// next int
// }

// func NewSelector() grpclb.Selector {
// 	return &Selector{
// 		hash:         consistent.New(),
// 		BaseSelector: *grpclb.NewBaseSelector(),
// 		logEntry: log.WithFields(
// 			log.Fields{
// 				"module": "rpc-balance",
// 			},
// 		),
// 	}
// }

// func (r *Selector) Add(addr grpc.Address) error {
// 	r.logEntry.Info("add host ", addr.Addr)
// 	r.hash.Add(addr.Addr)
// 	return r.BaseSelector.Add(addr)
// }

// func (r *Selector) Delete(addr grpc.Address) error {
// 	addr.Addr = addr.Addr[strings.LastIndex(addr.Addr, "/")+1:]
// 	r.logEntry.Info("remove host ", addr.Addr)
// 	r.hash.Remove(addr.Addr)
// 	return r.BaseSelector.Delete(addr)
// }

// func (r Selector) GetHostByClientID(clientID string) (string, error) {
// 	addr, err := r.hash.Get(clientID)
// 	if err != nil {
// 		return "", err
// 	}
// 	return addr[0:strings.Index(addr, ":")], err
// }

// func (r Selector) Get(ctx context.Context) (addr grpc.Address, err error) {
// 	_, clientID, _, err := CheckContext(ctx, false)
// 	if err != nil {
// 		return addr, err
// 	}

// 	// logEntry := r.
// 	// r.logEntry.Data["ClientID"] = clientID
// 	// r.logEntry.Data["RequestID"] = requestID
// 	// r.logEntry.Data["ClientID"] = clientID
// 	host, err := r.hash.Get(clientID)
// 	// fmt.Println()
// 	if err != nil {
// 		return addr, err
// 	}
// 	if addrInfo, ok := r.BaseSelector.GetAddrMap()[host]; ok {
// 		if addrInfo.Connected {
// 			addr = addrInfo.Addr
// 			// addrInfo.load++
// 			// r.next = next
// 			return
// 		}
// 	} else {
// 		err = fmt.Errorf("host:%v not found", host)
// 		return
// 	}
// 	// return r.BaseSelector.GetAddrMap()[host]
// 	// addr.Addr = host
// 	return
// }

// func GetRequestMetaDataWithOutgoingContext(ctx context.Context) metadata.MD {
// 	md, ok := metadata.FromOutgoingContext(ctx)
// 	if ok {
// 		return md
// 	}
// 	return nil
// }

// func GetRequestMetaDataWithIncomingContext(ctx context.Context) metadata.MD {
// 	md, ok := metadata.FromIncomingContext(ctx)
// 	if ok {
// 		return md
// 	}
// 	return nil
// }

// func CheckContext(ctx context.Context, in bool) (metadata.MD, string, string, error) {
// 	// logEntry := ctx.Value("logEntry")
// 	var md metadata.MD
// 	if in {
// 		md = GetRequestMetaDataWithIncomingContext(ctx)
// 	} else {
// 		md = GetRequestMetaDataWithOutgoingContext(ctx)
// 	}
// 	if md == nil {
// 		return nil, "", "", errors.New("context read error")
// 	}
// 	clientID := ""
// 	requestID := ""
// 	data, ok := md["clientid"]
// 	if !ok {
// 		// logEntry.Data["clientID"] = clientID
// 		return md, clientID, requestID, errors.New("context missing clientid")
// 	}
// 	clientID = data[0]

// 	data, ok = md["requestid"]
// 	if !ok {
// 		return md, clientID, requestID, errors.New("context missing requestid")
// 	}
// 	requestID = data[0]
// 	return md, clientID, requestID, nil

// }
