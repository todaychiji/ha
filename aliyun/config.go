package aliyun

type Config struct {
	AccessKeyID     string `validate:"required" env:"ACCESS_KEY"`
	AccessKeySecret string `validate:"required" env:"ACCESS_SECRET"`
	AccountID       string `validate:"required" env:"ACCOUNT_ID"`
	FcEndPoint      string `validate:"required" env:"FC_ENDPOINT,cn-beijing.fc.aliyuncs.com"`
	OssEndPoint     string `validate:"required" env:"OSS_ENDPOINT,oss-cn-beijing.aliyuncs.com"`
	OssBucketName   string `validate:"required" env:"OSS_BUCKET_NAME"`
	LogEndPoint     string `validate:"required" env:"LOG_ENDPOINT,cn-beijing.log.aliyuncs.com"`
}