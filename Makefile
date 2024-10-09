# cd to each directory prefixed with "ods-aurora" and run "go install ."
install:
	@for dir in $(shell ls -d myq-*/); do \
		cd $$dir && go install .; \
		cd ..; \
	done

test:
	@for dir in $(shell ls -d lib/); do \
		cd $$dir && go test ./...; \
		cd ..; \
	done
