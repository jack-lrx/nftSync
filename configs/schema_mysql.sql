-- NFT资产表
CREATE TABLE nfts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    token_id VARCHAR(128) NOT NULL,
    contract VARCHAR(128) NOT NULL,
    owner VARCHAR(128) NOT NULL,
    token_uri TEXT,
    metadata JSON,
    price VARCHAR(64),
    confidence INT DEFAULT 1,
    confirmed TINYINT(1) DEFAULT 0,
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
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    nft_id BIGINT NOT NULL,
    name VARCHAR(128),
    trait_type VARCHAR(64),
    value VARCHAR(128),
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT,
    FOREIGN KEY (nft_id) REFERENCES nfts(id) ON DELETE CASCADE
);
CREATE INDEX idx_items_nft_id ON items(nft_id);
CREATE INDEX idx_items_trait_type ON items(trait_type);

-- 挂单表（可选，示例）
CREATE TABLE orders (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    nft_id BIGINT NOT NULL,
    nft_token VARCHAR(128) NOT NULL,
    seller VARCHAR(128) NOT NULL,
    buyer VARCHAR(128),
    price VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL, -- listed, matched, completed, cancelled
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT,
    FOREIGN KEY (nft_id) REFERENCES nfts(id) ON DELETE CASCADE
);
CREATE INDEX idx_orders_nft_id ON orders(nft_id);
CREATE INDEX idx_orders_seller ON orders(seller);
CREATE INDEX idx_orders_buyer ON orders(buyer);
CREATE INDEX idx_orders_status ON orders(status);

-- 交易表（可选，示例）
CREATE TABLE trades (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT NOT NULL,
    buyer VARCHAR(128) NOT NULL,
    price VARCHAR(64) NOT NULL,
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);
CREATE INDEX idx_trades_order_id ON trades(order_id);
CREATE INDEX idx_trades_buyer ON trades(buyer);
