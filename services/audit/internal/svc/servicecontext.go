package svc

import (
	"time"

	"github.com/avast/retry-go"
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
	auditMq, err := initAuditMQ(c)
	if err != nil {
		logx.Errorf("audit mq init failed after retries: %v", err)
		panic(err)
	}
	return &ServiceContext{
		Config:  c,
		AuditMQ: auditMq,
	}
}

func initAuditMQ(c config.Config) (*mq.AuditMQ, error) {
	var (
		auditMQ *mq.AuditMQ
		err     error
	)

	retryErr := retry.Do(
		func() error {
			auditMQ, err = mq.Init(c)
			return err
		},
		retry.Attempts(30),
		retry.Delay(2*time.Second),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			logx.Errorf("audit mq init attempt %d/30 failed: %v", n+1, err)
		}),
	)
	if retryErr != nil {
		return nil, retryErr
	}

	return auditMQ, nil
}
