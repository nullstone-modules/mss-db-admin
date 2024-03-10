NAME := mss-db-admin

.PHONY: tools build

tools:
	cd ~ && go get -u github.com/aws/aws-lambda-go/cmd/build-lambda-zip && cd -

build:
	mkdir -p ./aws/tf/files
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o ./aws/tf/files/bootstrap ./aws/

package: tools
	cd ./aws/tf \
		&& build-lambda-zip --output files/mss-db-admin.zip files/bootstrap \
		&& tar -cvzf aws-module.tgz *.tf files/mss-db-admin.zip \
		&& mv aws-module.tgz ../../

acc: acc-up acc-run acc-down

acc-up:
	cd acc && docker-compose -p mss-db-admin-acc up -d db

acc-run:
	ACC=1 gotestsum ./acc/...

acc-down:
	cd acc && docker-compose -p mss-db-admin-acc down
