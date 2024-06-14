GO_ENV := GOOS=linux GOARCH=amd64
# GO_ENV := $env:GOOS="linux"; $env:GOARCH="amd64"

# BUILD_FLAGS := -ldflags="-s -w"
BUILD_FLAGS := -ldflags="-s -w" -tags lambda.norpc

build:
	$(GO_ENV) go build $(BUILD_FLAGS) -o bin/register/bootstrap pkg/user/create_user/register.go
	(cd bin/register && zip -FS bootstrap.zip bootstrap)
	$(GO_ENV) go build $(BUILD_FLAGS) -o bin/deleteuser/bootstrap pkg/user/delete_user/delete.go
	(cd bin/deleteuser && zip -FS bootstrap.zip bootstrap)
	$(GO_ENV) go build $(BUILD_FLAGS) -o bin/getuser/bootstrap pkg/user/get_user/getuser.go
	(cd bin/getuser && zip -FS bootstrap.zip bootstrap)
	$(GO_ENV) go build $(BUILD_FLAGS) -o bin/updateuser/bootstrap pkg/user/update_user/update.go
	(cd bin/updateuser && zip -FS bootstrap.zip bootstrap)
	$(GO_ENV) go build $(BUILD_FLAGS) -o bin/listuser/bootstrap pkg/user/list_user/listuser.go
	(cd bin/listuser && zip -FS bootstrap.zip bootstrap)
	$(GO_ENV) go build $(BUILD_FLAGS) -o bin/login/bootstrap pkg/login/login.go
	(cd bin/login && zip -FS bootstrap.zip bootstrap)
	$(GO_ENV) go build $(BUILD_FLAGS) -o bin/authentication/bootstrap pkg/authentication/authentication.go
	(cd bin/authentication && zip -FS bootstrap.zip bootstrap)
	$(GO_ENV) go build $(BUILD_FLAGS) -o bin/redirecturl/bootstrap pkg/redirect_url/redirect.go
	(cd bin/redirecturl && zip -FS bootstrap.zip bootstrap)
	$(GO_ENV) go build $(BUILD_FLAGS) -o bin/shorturl/bootstrap pkg/shorten_url/short_create.go
	(cd bin/shorturl && zip -FS bootstrap.zip bootstrap)
	# tracking
	$(GO_ENV) go build $(BUILD_FLAGS) -o bin/trackbyurl/bootstrap pkg/analyticts/track_shorturl/track_url.go
	(cd bin/trackbyurl && zip -FS bootstrap.zip bootstrap)
	$(GO_ENV) go build $(BUILD_FLAGS) -o bin/trackbyuser/bootstrap pkg/analyticts/track_by_user/trackbyuser.go
	(cd bin/trackbyuser && zip -FS bootstrap.zip bootstrap)
	# recovery
	$(GO_ENV) go build $(BUILD_FLAGS) -o bin/passrecovery/bootstrap pkg/recovery/pass_recovery/pass_recovery.go
	(cd bin/passrecovery && zip -FS bootstrap.zip bootstrap)
	$(GO_ENV) go build $(BUILD_FLAGS) -o bin/consumer/bootstrap pkg/recovery/email_consumer/sqs_consumer.go
	(cd bin/consumer && zip -FS bootstrap.zip bootstrap)


clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy:
	serverless deploy --stage prod --verbose

