-- Down migration for 000255_finalize_unified_wallet_migration
-- This reverses the data migration (not the schema - schema is reversed by 000254)

BEGIN;

-- Remove migrated wallet transactions (from fee_transactions)
DELETE FROM wallet_transactions
WHERE id IN (
    SELECT ft.id
    FROM fee_transactions ft
    JOIN wallets w ON w.user_id = ft.user_id AND w.owner_type = 'user'
);

-- Remove migrated wallet transactions (from agent_financial_transactions)
DELETE FROM wallet_transactions
WHERE id IN (
    SELECT aft.id
    FROM agent_financial_transactions aft
    JOIN wallets w ON w.agent_id = aft.to_agent_id AND w.owner_type = 'agent'
);

-- Remove migrated agent wallets
DELETE FROM wallets
WHERE owner_type = 'agent'
AND id IN (
    SELECT gen_random_uuid() AS id
    FROM agent_billing_controls abc
    WHERE NOT EXISTS (
        SELECT 1 FROM wallets w
        WHERE w.owner_type = 'agent' AND w.owner_id = abc.agent_id
    )
);

-- Remove migrated user wallets
DELETE FROM wallets
WHERE owner_type = 'user'
AND user_id IN (SELECT user_id FROM user_wallets);

COMMIT;