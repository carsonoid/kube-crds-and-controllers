DIFF_REPO      ?= kube-crds-and-controllers-diffs
DIFF_REPO_PATH  = ../$(DIFF_REPO)
DIFF_REPO_URL   = git@github.com:carsonoid/kube-crds-and-controllers-diffs.git
DIFF_REPO_GIT   = git -C $(DIFF_REPO_PATH)

PHONY: deps clean gen-go-crds diffs-repo push-diffs-repo 

deps:
	glide i

clean: clean-build clean-go-crds

clean-build:
	# Clean up build dir
	rm -rf build/*

# Example Controllers

clean-go-crds:
	# Delete all generated code.
	rm -rf controllers/crd-configured/pkg/client
	rm -f  controllers/crd-configured/pkg/apis/*/*/zz_generated.deepcopy.go

gen-go-crds: clean-go-crds # Must be cleaned so that removals happen properly
	./vendor/k8s.io/code-generator/generate-groups.sh \
	all \
	github.com/carsonoid/kube-crds-and-controllers/controllers/crd-configured/pkg/client \
	github.com/carsonoid/kube-crds-and-controllers/controllers/crd-configured/pkg/apis \
	"podlabeler:v1alpha1"

	# workaround https://github.com/openshift/origin/issues/10357
	find controllers/*/pkg/client -name "clientset_generated.go" -exec sed -i'' 's/return \\&Clientset{fakePtr/return \\&Clientset{\\&fakePtr/g' '{}' \;

controllers/%:
	@mkdir build >/dev/null 2>&1|| true
	go build -i -o build/$@ $@.go

run-controllers/%: controllers/%
	./build/controllers/$* $(OPTS)

# Represent controller revisions in a single git repo commit set.
diffs-repo:
	rm -rf $(DIFF_REPO_PATH) || true
	mkdir $(DIFF_REPO_PATH)
	$(DIFF_REPO_GIT) init
	$(DIFF_REPO_GIT) remote add origin $(DIFF_REPO_URL)
	echo "# Kube Controller Evolution" > $(DIFF_REPO_PATH)/README.md
	echo >> $(DIFF_REPO_PATH)/README.md
	echo "Diff view of controllers from original repo: https://github.com/carsonoid/kube-crds-and-controllers" >> $(DIFF_REPO_PATH)/README.md
	$(DIFF_REPO_GIT) add README.md
	cp LICENSE $(DIFF_REPO_PATH)/
	$(DIFF_REPO_GIT) add LICENSE
	$(DIFF_REPO_GIT) commit -m "Initial commit"
	for c in \
	controllers/hard-coded/simple.go \
	controllers/hard-coded/structured.go \
	controllers/configmap-configured/single-config.go \
	controllers/configmap-configured/multi-config.go \
	controllers/crd-configured/simple.go \
	controllers/crd-configured/workqueue.go \
	; do \
	cp $$c $(DIFF_REPO_PATH)/controller.go && \
	$(DIFF_REPO_GIT) add controller.go &&\
	$(DIFF_REPO_GIT) commit -m "Move to $$c" \
	; done
	# Don't force push to master
	$(DIFF_REPO_GIT) checkout -b dev

push-diffs-repo: diffs-repo
	$(DIFF_REPO_GIT) push -f -u origin dev

# Workshop Provisioner

wp-clean-go-crds:
	# Delete all generated code.
	rm -rf controllers/workshop-provisioner/pkg/client
	rm -f  controllers/workshop-provisioner/pkg/apis/*/*/zz_generated.deepcopy.go

wp-gen-go-crds: wp-clean-go-crds # Must be cleaned so that removals happen properly
	./vendor/k8s.io/code-generator/generate-groups.sh \
	all \
	github.com/carsonoid/kube-crds-and-controllers/controllers/workshop-provisioner/pkg/client \
	github.com/carsonoid/kube-crds-and-controllers/controllers/workshop-provisioner/pkg/apis \
	"provisioner:v1alpha1"

workshop-provisioner-all: workshop-provisioner-crds workshop-provisioner

workshop-provisioner-crds: wp-clean-go-crds wp-gen-go-crds

workshop-provisioner: controllers/workshop-provisioner/workshop-provisioner

run-provisioner: run-controllers/workshop-provisioner/workshop-provisioner
