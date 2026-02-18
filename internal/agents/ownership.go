package agents

import "context"

type contextKey string

const agentCtxKey contextKey = "agent"

func SetAgentInContext(ctx context.Context, agent *Agent) context.Context {
	return context.WithValue(ctx, agentCtxKey, agent)
}

func GetAgentFromContext(ctx context.Context) *Agent {
	agent, _ := ctx.Value(agentCtxKey).(*Agent)
	return agent
}
