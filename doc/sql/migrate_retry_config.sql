use `go-mco-gateway`;

-- 新增 retry_config 字段，存储重试策略配置（JSON），与 params 字段同风格
ALTER TABLE tools ADD COLUMN retry_config JSON AFTER params;
