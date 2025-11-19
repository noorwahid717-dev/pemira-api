package constants

type Role string

const (
	RoleStudent      Role = "STUDENT"
	RoleAdmin        Role = "ADMIN"
	RoleTPSOperator  Role = "TPS_OPERATOR"
	RoleSuperAdmin   Role = "SUPER_ADMIN"
)

type ElectionPhase string

const (
	PhaseRegistration ElectionPhase = "REGISTRATION"
	PhaseCampaign     ElectionPhase = "CAMPAIGN"
	PhaseVoting       ElectionPhase = "VOTING"
	PhaseCounting     ElectionPhase = "COUNTING"
	PhaseAnnouncement ElectionPhase = "ANNOUNCEMENT"
)

type VotingMode string

const (
	VotingModeOnline VotingMode = "ONLINE"
	VotingModeTPS    VotingMode = "TPS"
	VotingModeHybrid VotingMode = "HYBRID"
)

type VoterStatus string

const (
	VoterStatusEligible   VoterStatus = "ELIGIBLE"
	VoterStatusVoted      VoterStatus = "VOTED"
	VoterStatusIneligible VoterStatus = "INELIGIBLE"
)
