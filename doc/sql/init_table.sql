use `go-mco-gateway`;

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

