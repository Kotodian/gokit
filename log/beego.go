package log

import (
	_log "github.com/sirupsen/logrus"
	"github.com/astaxie/beego/context"
	// "github.com/astaxie/beego/context/param"
)

func FromBeegoContext(ctx *context.Context) *_log.Entry {
	// func GetLogFromContext(ctx *context.Context) *log.Entry {
	entry := ctx.Input.GetData("logEntry")
	if entry == nil {
		_entry := _log.WithFields(_log.Fields{
			"module":    "admin",
			"requestID": ctx.Input.GetData("requestID"),
		})
		ctx.Input.SetData("logEntry", _entry)
		return _entry
	}
	return entry.(*_log.Entry)
}

func NewFromBeegoContext(ctx *context.Context, module string) *_log.Entry {
	oldLogEntry := FromBeegoContext(ctx)
	entry := _log.WithFields(_log.Fields{
		"module":    module,
		"requestID": oldLogEntry.Data["requestID"],
	})
	ctx.Input.SetData("logEntry", entry)
	return entry
}
