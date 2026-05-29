-- 测试数据
use `go-mco-gateway`;

-- 下游服务
INSERT INTO projects (project_id, name, base_url, description)
VALUES ('proj_user', '用户服务', 'https://jsonplaceholder.typicode.com', '用户相关的 HTTP 服务');

INSERT INTO projects (project_id, name, base_url, description)
VALUES ('proj_post', '帖子服务', 'https://jsonplaceholder.typicode.com', '帖子相关的 HTTP 服务');

-- 工具定义
INSERT INTO tools (project_id, name, title, description, http_method, url_template, params)
VALUES (
    'proj_user',
    'get_user',
    '获取用户信息',
    '根据用户 ID 获取用户的基本信息，包括姓名、邮箱、电话等。可指定返回字段。',
    'GET',
    '/users/{user_id}',
    '[
        {"name":"user_id","type":"number","location":"path","required":true,"description":"用户 ID"},
        {"name":"fields","type":"string","location":"query","required":false,"default_value":"","description":"返回字段，逗号分隔"}
    ]'
);

INSERT INTO tools (project_id, name, title, description, http_method, url_template, params)
VALUES (
    'proj_user',
    'get_user_posts',
    '获取用户帖子',
    '根据用户 ID 获取该用户的所有帖子列表。',
    'GET',
    '/users/{user_id}/posts',
    '[
        {"name":"user_id","type":"number","location":"path","required":true,"description":"用户 ID"}
    ]'
);

INSERT INTO tools (project_id, name, title, description, http_method, url_template, params)
VALUES (
    'proj_post',
    'create_post',
    '创建帖子',
    '创建一个新的帖子，需要提供标题和内容。',
    'POST',
    '/posts',
    '[
        {"name":"title","type":"string","location":"body","required":true,"description":"帖子标题"},
        {"name":"body","type":"string","location":"body","required":true,"description":"帖子内容"},
        {"name":"userId","type":"number","location":"body","required":true,"description":"用户 ID"}
    ]'
);
