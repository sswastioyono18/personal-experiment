insert into public.casbin_rule_test (p_type, v0, v1, v2, v3, v4, v5)
values  ('g', 'tio', 'admin', '', '', '', ''),
        ('g', 'tiopartner', 'partner', '', '', '', ''),
        ('g', 'tiosuperadmin', 'superadmin', '', '', '', ''),
        ('p', 'admin', '/api/v1/admin/*', 'GET|POST', '', '', ''),
        ('p', 'superadmin', '/api/v1/superadmin/*', 'GET|POST', '', '', ''),
        ('p', 'partner', '/api/v1/partner/*', 'GET|POST|PUT', '', '', '');