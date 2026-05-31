use `go-mco-gateway`;

-- 新增 retry_config 字段，存储重试策略配置（JSON），与 params 字段同风格
ALTER TABLE tools ADD COLUMN retry_config JSON AFTER params;

-- 添加限流配置字段
ALTER TABLE tools
    ADD COLUMN rate_limit_config JSON DEFAULT NULL
        COMMENT '限流配置: {"max_requests":100,"window_seconds":1}, null=不限流';
