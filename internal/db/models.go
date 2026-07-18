package db

import "time"

type Session struct {
	ID, Title, Directory, Agent, ModelID string
	Cost                                 float64
	TokensIn, TokensOut, TokensReason    int64
	Created, Updated                     time.Time
	Compacting                           *time.Time
}

type Part struct {
	ID, Type, Tool, Text string
	HasOutput            bool
	Created              time.Time
}

type AgentRow struct {
	Agent, Model string
	Count        int
	Cost         float64
	Tokens       int64
	LastActive   time.Time
}

type ProjectRow struct {
	Dir, FullPath, Agent, Model string
	Count                       int
	Cost                        float64
	Tokens                      int64
	LastActive                  time.Time
}

type PeriodRow struct {
	Label  string
	Cost   float64
	Tokens int64
}

type Activity struct {
	SessionID string
	Type      string
	Tool      string
	HasOutput bool
	Created   time.Time
}

type SearchResult struct {
	SessionID, SessionTitle, PartType, Text string
}
