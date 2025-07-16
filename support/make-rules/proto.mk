# 设置变量
PROTO_DIR := .proto
PROTOC_DIR := $(PROTO_DIR)/protoc

# protoc 26.1
PROTOC_FILE_NAME := protoc-26.1-win64.zip
PROTOC_URL := https://cdn-plaso-school.plaso.cn/cdn/proto/$(PROTOC_FILE_NAME)
PROTOC_BIN := $(PROTOC_DIR)/bin/protoc.exe

# kratos 2.7.3 整合包
PROTOC_ZIP_NAME := proto_bin.zip
PROTOC_ZIP_URL := https://cdn-plaso-school.plaso.cn/cdn/proto/$(PROTOC_ZIP_NAME)

KRATOS_BIN := $(PROTO_DIR)/kratos.exe




.PHONY: proto-init-env
proto-init-env:
# 设置环境变量
ENVPATH := $(shell echo $$PATH)
export PATH=${CURRENT_DIR}/${PROTO_DIR}:${CURRENT_DIR}/$(PROTOC_DIR)/bin:$(ENVPATH)

.PHONY: proto-init
proto-init:
# 检查文件夹是否存在
ifeq ($(wildcard $(PROTO_DIR)),)
	@echo "Creating $(PROTO_DIR) directory..."
	@mkdir -p $(PROTO_DIR)
endif

ifeq ($(wildcard $(PROTOC_DIR)),)
	@echo "Creating $(PROTOC_DIR) directory..."
	@mkdir -p $(PROTOC_DIR)
endif


# 下载protoc
ifeq ($(wildcard $(PROTOC_BIN)),)
	@echo "protoc not found. Downloading..."
	@curl -L $(PROTOC_URL) -o $(PROTO_DIR)/$(PROTOC_FILE_NAME)
	# 解压
	@unzip $(PROTO_DIR)/$(PROTOC_FILE_NAME) -d $(PROTOC_DIR)
	@echo "protoc-gen-go downloaded and made executable."
endif

# 下载kratos
ifeq ($(wildcard $(KRATOS_BIN)),)
	@echo "kratos not found. Downloading..."
	curl -L $(PROTOC_ZIP_URL) -o $(PROTO_DIR)/$(PROTOC_ZIP_NAME)
	@unzip $(PROTO_DIR)/$(PROTOC_ZIP_NAME) -d $(PROTO_DIR)
	@echo "kratos downloaded and made executable."
endif



