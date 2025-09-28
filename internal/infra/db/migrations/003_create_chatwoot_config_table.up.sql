-- Create chatwoot_config table
CREATE TABLE IF NOT EXISTS "zpChatwoot" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "sessionId" UUID NOT NULL REFERENCES "zpSessions"("id") ON DELETE CASCADE,
    "url" VARCHAR(2048) NOT NULL,
    "token" VARCHAR(255) NOT NULL,
    "accountId" VARCHAR(50) NOT NULL,
    "inboxId" VARCHAR(50),
    "enabled" BOOLEAN NOT NULL DEFAULT true,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_session_id" ON "zpChatwoot" ("sessionId");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_enabled" ON "zpChatwoot" ("enabled");
CREATE INDEX IF NOT EXISTS "idx_zp_chatwoot_created_at" ON "zpChatwoot" ("createdAt");

-- Unique constraint: one Chatwoot config per session
CREATE UNIQUE INDEX IF NOT EXISTS "idx_zp_chatwoot_unique_session" ON "zpChatwoot" ("sessionId");

-- Create trigger to automatically update updatedAt
CREATE TRIGGER update_zp_chatwoot_updated_at
    BEFORE UPDATE ON "zpChatwoot"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE "zpChatwoot" IS 'Chatwoot integration configuration - one per session';
COMMENT ON COLUMN "zpChatwoot"."id" IS 'Unique configuration identifier';
COMMENT ON COLUMN "zpChatwoot"."sessionId" IS 'Reference to WhatsApp session (one-to-one)';
COMMENT ON COLUMN "zpChatwoot"."url" IS 'Chatwoot instance URL';
COMMENT ON COLUMN "zpChatwoot"."token" IS 'Chatwoot user token';
COMMENT ON COLUMN "zpChatwoot"."accountId" IS 'Chatwoot account ID';
COMMENT ON COLUMN "zpChatwoot"."inboxId" IS 'Optional Chatwoot inbox ID';
COMMENT ON COLUMN "zpChatwoot"."enabled" IS 'Whether configuration is enabled';
COMMENT ON COLUMN "zpChatwoot"."createdAt" IS 'Configuration creation timestamp';
COMMENT ON COLUMN "zpChatwoot"."updatedAt" IS 'Last update timestamp';
