-- Emergency migration to fix admin player constraint violation
-- Run this directly against your database before deploying

BEGIN;

-- Gather admin player_ids and clean up artifacts
DO $$
DECLARE
    admin_player_ids TEXT[];
BEGIN
    SELECT ARRAY_AGG(DISTINCT player_id)
    INTO admin_player_ids
    FROM accounts
    WHERE role IN ('admin', 'frozen:admin')
      AND player_id IS NOT NULL;

    IF admin_player_ids IS NOT NULL THEN
        -- Delete from all player history tables
        DELETE FROM coin_earning_log WHERE player_id = ANY(admin_player_ids);
        DELETE FROM star_purchase_log WHERE player_id = ANY(admin_player_ids);
        
        -- Conditionally delete from optional tables
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'player_faucet_claims') THEN
            DELETE FROM player_faucet_claims WHERE player_id = ANY(admin_player_ids);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'player_star_variants') THEN
            DELETE FROM player_star_variants WHERE player_id = ANY(admin_player_ids);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'player_boosts') THEN
            DELETE FROM player_boosts WHERE player_id = ANY(admin_player_ids);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'player_ip_associations') THEN
            DELETE FROM player_ip_associations WHERE player_id = ANY(admin_player_ids);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'player_telemetry') THEN
            DELETE FROM player_telemetry WHERE player_id = ANY(admin_player_ids);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'bug_reports') THEN
            DELETE FROM bug_reports WHERE player_id = ANY(admin_player_ids);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tsa_cinder_sigils') THEN
            DELETE FROM tsa_cinder_sigils WHERE owner_player_id = ANY(admin_player_ids);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tsa_mint_log') THEN
            DELETE FROM tsa_mint_log WHERE buyer_player_id = ANY(admin_player_ids);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tsa_trade_log') THEN
            DELETE FROM tsa_trade_log WHERE seller_player_id = ANY(admin_player_ids) OR buyer_player_id = ANY(admin_player_ids);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tsa_activation_log') THEN
            DELETE FROM tsa_activation_log WHERE player_id = ANY(admin_player_ids);
        END IF;
        
        -- Delete player rows
        DELETE FROM players WHERE player_id = ANY(admin_player_ids);
    END IF;

    -- Null out player_id for admin accounts
    UPDATE accounts SET player_id = NULL WHERE role IN ('admin', 'frozen:admin');
END $$;

-- Now the constraint can be safely added
ALTER TABLE accounts ALTER COLUMN player_id DROP NOT NULL;
ALTER TABLE accounts DROP CONSTRAINT IF EXISTS accounts_admin_player_check;
ALTER TABLE accounts ADD CONSTRAINT accounts_admin_player_check
    CHECK (
        (role IN ('admin', 'frozen:admin') AND player_id IS NULL)
        OR (role NOT IN ('admin', 'frozen:admin') AND player_id IS NOT NULL)
    );

COMMIT;
