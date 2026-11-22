CREATE TABLE widget_components (
    widget_id text    NOT NULL REFERENCES widgets (id) ON DELETE CASCADE,
    position  integer NOT NULL CHECK (position >= 0),
    type      text    NOT NULL CHECK (type IN ('INFORMATION', 'PRODUCT', 'CATEGORY')),
    sub_type  text    NOT NULL DEFAULT '',
    image_url text    NOT NULL DEFAULT '',
    is_active boolean NOT NULL,
    data_src  text    NOT NULL DEFAULT '',
    PRIMARY KEY (widget_id, position)
);
