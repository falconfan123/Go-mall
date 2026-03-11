package svc

import (
	"github.com/falconfan123/Go-mall/dal/model/audit"
	"github.com/falconfan123/Go-mall/services/audit/internal/config"
	"github.com/falconfan123/Go-mall/services/audit/internal/mq"
	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config     config.Config
	AuditMQ    *mq.AuditMQ
	AuditModel audit.AuditModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	auditMq, err := mq.Init(c)
	if err != nil {
		logx.Error(err)
		panic(err)
	}
	return &ServiceContext{
		Config:  c,
		AuditMQ: auditMq,
	}
}
