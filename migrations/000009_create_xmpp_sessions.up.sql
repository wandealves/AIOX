CREATE TABLE xmpp_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id),
    owner_user_id UUID NOT NULL REFERENCES users(id),
    jid TEXT NOT NULL,
    from_jid TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ
);

CREATE INDEX idx_xmpp_sessions_agent_id ON xmpp_sessions(agent_id);
CREATE INDEX idx_xmpp_sessions_owner ON xmpp_sessions(owner_user_id);
CREATE INDEX idx_xmpp_sessions_status ON xmpp_sessions(status) WHERE status = 'active';
