all: install build
.PHONY: all

LDFLAGS = -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.branch=${BRANCH}"
include Makefile.Inc

install:
.PHONY: install

build: $(BINDIR_SVM_LIBS)

.PHONY: build

SVM_DIR=$(PROJ_DIR)../svm/
build-rust-svm:
	mkdir -p $(BIN_DIR) svm/wasm
	cd $(SVM_DIR) && cargo +nightly build -p svm-runtime-ffi --release \
	$(foreach X,$(SVM_LIBS),&& cp $(SVM_DIR)target/release/$(X) $(BIN_DIR)$(X))
	cp $(SVM_DIR)target/release/svm.h svm/
	cp $(SVM_DIR)crates/runtime-ffi/tests/wasm/counter.wasm svm/wasm/
	cp $(SVM_DIR)crates/runtime-ffi/tests/wasm/failure.wasm svm/wasm/
.PHONY: build-rust-svm
