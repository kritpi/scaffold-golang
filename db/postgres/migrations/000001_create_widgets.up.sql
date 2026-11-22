CREATE TABLE widgets (
    id                  text        PRIMARY KEY,
    header_th           text        NOT NULL CHECK (header_th <> ''),
    header_en           text        NOT NULL DEFAULT '',
    sub_header_th       text        NOT NULL DEFAULT '',
    sub_header_en       text        NOT NULL DEFAULT '',
    start_date          timestamptz NOT NULL,
    end_date            timestamptz NOT NULL,
    product_list_ref_id text        NOT NULL DEFAULT '',
    created_at          timestamptz NOT NULL,
    created_by          text        NOT NULL,
    updated_at          timestamptz NOT NULL,
    updated_by          text        NOT NULL,
    CONSTRAINT widgets_date_order CHECK (end_date >= start_date)
);

CREATE INDEX widgets_updated_at_id_idx ON widgets (updated_at DESC, id);
