gen-go-crs:
	./vendor/k8s.io/code-generator/generate-groups.sh \
	all \
	github.com/carsonoid/kube-crds-and-controllers/crd-configured-controller/pkg/client \
	github.com/carsonoid/kube-crds-and-controllers/crd-configured-controller/pkg/apis \
	"podlabelconfig:v1alpha1"

	# workaround https://github.com/openshift/origin/issues/10357
	find pkg/client -name "clientset_generated.go" -exec sed -i'' 's/return \\&Clientset{fakePtr/return \\&Clientset{\\&fakePtr/g' '{}' \;

all: clean godeps gen-go-crs test

godeps:
	glide i

clean:
	# Delete all generated code.
	rm -rf pkg/client
	rm -f pkg/apis/*/*/zz_generated.deepcopy.go

# # This is a basic smoke-test to make sure the types compile
# test:
# 	go build -o build/crud -i _examples/clients/crud.go
# 	go build -o build/list -i _examples/clients/list.go

# 	@echo "All Test binaries Compiled!"
