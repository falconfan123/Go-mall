package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/falconfan123/Go-mall/services/product/internal/svc"
	"github.com/falconfan123/Go-mall/services/product/product"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUploadURLLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUploadURLLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUploadURLLogic {
	return &GetUploadURLLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MinIO 预签名上传接口
func (l *GetUploadURLLogic) GetUploadURL(in *product.GetUploadURLReq) (*product.GetUploadURLResp, error) {
	// 1. Define bucket and key
	bucketName := l.svcCtx.Config.Minio.Bucket
	objectName := fmt.Sprintf("products/%d/%s.jpg", time.Now().Year(), uuid.New().String())

	// 2. Set policy conditions
	policy := minio.NewPostPolicy()
	policy.SetBucket(bucketName)
	policy.SetKey(objectName)
	policy.SetExpires(time.Now().UTC().Add(15 * time.Minute))
	policy.SetContentLengthRange(0, 5*1024*1024) // 0 - 5MB

	// Optional: Restrict content type if provided
	if in.ContentType != "" {
		policy.SetContentType(in.ContentType)
	}

	// 3. Generate presigned URL and form data
	urlStr, formData, err := l.svcCtx.MinioClient.PresignedPostPolicy(l.ctx, policy)
	if err != nil {
		l.Logger.Errorw("failed to generate presigned policy", logx.Field("err", err))
		return &product.GetUploadURLResp{
			StatusCode: 500,
			StatusMsg:  "failed to generate upload url",
		}, nil
	}

	return &product.GetUploadURLResp{
		UploadUrl:  urlStr.String(),
		Key:        objectName,
		FormData:   formData,
		StatusCode: 0,
		StatusMsg:  "success",
	}, nil
}
