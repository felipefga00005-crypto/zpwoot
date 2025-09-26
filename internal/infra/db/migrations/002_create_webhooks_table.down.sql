-- Drop webhooks table
DROP TRIGGER IF EXISTS update_zp_webhooks_updated_at ON "zpWebhooks";
DROP TABLE IF EXISTS "zpWebhooks";
