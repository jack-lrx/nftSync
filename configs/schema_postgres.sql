-- NFT资产表
CREATE TABLE nfts (
    id SERIAL PRIMARY KEY,
    token_id VARCHAR(128) NOT NULL,
    contract VARCHAR(128) NOT NULL,
    owner VARCHAR(128) NOT NULL,
    token_uri TEXT,
    metadata JSONB,
    price VARCHAR(64),
    confidence INT DEFAULT 1,
    confirmed BOOLEAN DEFAULT FALSE,
    source_nodes TEXT,
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT
);
CREATE INDEX idx_nfts_token_id_contract ON nfts(token_id, contract);
CREATE INDEX idx_nfts_owner ON nfts(owner);
CREATE INDEX idx_nfts_confirmed ON nfts(confirmed);

-- NFT属性表
CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    nft_id INT NOT NULL REFERENCES nfts(id) ON DELETE CASCADE,
    name VARCHAR(128),
    trait_type VARCHAR(64),
    value VARCHAR(128),
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT
);
CREATE INDEX idx_items_nft_id ON items(nft_id);
CREATE INDEX idx_items_trait_type ON items(trait_type);

-- 挂单表（可选，示例）
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    nft_id INT NOT NULL REFERENCES nfts(id) ON DELETE CASCADE,
    seller VARCHAR(128) NOT NULL,
    price VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL, -- pending, completed, cancelled
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT
);
CREATE INDEX idx_orders_nft_id ON orders(nft_id);
CREATE INDEX idx_orders_seller ON orders(seller);
CREATE INDEX idx_orders_status ON orders(status);

-- 交易表（可选，示例）
CREATE TABLE trades (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    buyer VARCHAR(128) NOT NULL,
    price VARCHAR(64) NOT NULL,
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT
);
CREATE INDEX idx_trades_order_id ON trades(order_id);
CREATE INDEX idx_trades_buyer ON trades(buyer);

