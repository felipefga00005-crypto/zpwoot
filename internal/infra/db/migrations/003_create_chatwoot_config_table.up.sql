-- Create chatwoot_config table
CREATE TABLE IF NOT EXISTS "zpChatwoot" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "url" VARCHAR(2048) NOT NULL,
    "token" VARCHAR(255) NOT NULL,
    "accountId" VARCHAR(50) NOT NULL,
    "inboxId" VARCHAR(50),
    "active" BOOLEAN NOT NULL DEFAULT true,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_active" ON "zpChatwoot" ("active");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_created_at" ON "zpChatwoot" ("createdAt");

-- Create trigger to automatically update updatedAt
CREATE TRIGGER update_zp_chatwoot_updated_at
    BEFORE UPDATE ON "zpChatwoot"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE "zpChatwoot" IS 'Chatwoot integration configuration';
COMMENT ON COLUMN "zpChatwoot"."id" IS 'Unique configuration identifier';
COMMENT ON COLUMN "zpChatwoot"."url" IS 'Chatwoot instance URL';
COMMENT ON COLUMN "zpChatwoot"."token" IS 'Chatwoot user token';
COMMENT ON COLUMN "zpChatwoot"."accountId" IS 'Chatwoot account ID';
COMMENT ON COLUMN "zpChatwoot"."inboxId" IS 'Optional Chatwoot inbox ID';
COMMENT ON COLUMN "zpChatwoot"."active" IS 'Whether configuration is active';
COMMENT ON COLUMN "zpChatwoot"."createdAt" IS 'Configuration creation timestamp';
COMMENT ON COLUMN "zpChatwoot"."updatedAt" IS 'Last update timestamp';
