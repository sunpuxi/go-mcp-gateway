CREATE TABLE projects (
                          project_id     VARCHAR(64) PRIMARY KEY,
                          name           VARCHAR(128) NOT NULL,
                          base_url       VARCHAR(512) NOT NULL,
                          description    TEXT,
                          status         TINYINT NOT NULL DEFAULT 1,
                          created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                          updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE tools (
                       tool_id        BIGINT PRIMARY KEY AUTO_INCREMENT,
                       project_id     VARCHAR(64) NOT NULL,
                       name           VARCHAR(128) NOT NULL,
                       title          VARCHAR(256),
                       description    TEXT,
                       http_method    VARCHAR(10) NOT NULL DEFAULT 'GET',
                       url_template   VARCHAR(512) NOT NULL,
                       timeout_ms     INT NOT NULL DEFAULT 5000,
                       params         JSON,                              -- ← 参数映射规则直接放这里
                       status         TINYINT NOT NULL DEFAULT 1,
                       created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                       UNIQUE KEY uk_name (name),
                       KEY idx_project_id (project_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE clients (
                         client_id      VARCHAR(64)  PRIMARY KEY,    -- 唯一标识，如 cli_bigdata
                         name           VARCHAR(128) NOT NULL,        -- 显示名称，如"大数据项目组"
                         api_key_hash   VARCHAR(64)  NOT NULL,        -- SHA256(完整 api_key)
                         api_key_prefix VARCHAR(32)  NOT NULL,        -- 前缀，如 sk-bigdata，用于管理后台展示
                         description    TEXT,                         -- 备注说明
                         status         TINYINT      NOT NULL DEFAULT 1,  -- 1=启用, 0=禁用
                         created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
                         updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                         INDEX idx_hash (api_key_hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- client_tool_permissions 改为 tool_id 关联
CREATE TABLE client_tool_permissions (
                                         id         BIGINT       AUTO_INCREMENT PRIMARY KEY,
                                         client_id  VARCHAR(64)  NOT NULL,
                                         tool_id    BIGINT       NOT NULL,            -- 改为关联 tools.tool_id
                                         created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,

                                         UNIQUE KEY uk_client_tool (client_id, tool_id),
                                         KEY idx_client (client_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


