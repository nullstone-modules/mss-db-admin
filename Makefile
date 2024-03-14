NAME := mss-db-admin

.PHONY: tools build

tools:
	go install github.com/aws/aws-lambda-go/cmd/build-lambda-zip@latest

build:
	mkdir -p ./aws/tf/files
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o ./aws/tf/files/bootstrap ./aws/

package: tools
	# Package aws module using build-lambda-zip which produces a viable package from any OS
	cd ./aws/tf && build-lambda-zip --output files/mss-db-admin.zip files/bootstrap

acc: acc-up acc-run acc-down

acc-up:
	cd acc && docker-compose -p mss-db-admin-acc up -d db

acc-run:
	ACC=1 gotestsum ./acc/...

acc-down:
	cd acc && docker-compose -p mss-db-admin-acc down
