package model

const (
	PROFILE_STATUS_ACTIVE          = "active"
	PROFILE_STATUS_LIQUIDATING     = "liquidating"
	VAULT_STATUS_ACTIVE            = "active"
	VAULT_STATUS_SUSPENDED         = "suspended"
	PROFILE_TYPE_TRADER            = "trader"
	PROFILE_TYPE_VAULT             = "vault"
	PROFILE_TYPE_INSURANCE         = "insurance"
	PROFILE_NOT_FOUND_ERROR        = "PROFILE_NOT_FOUND"
	LONG                           = "long"
	SHORT                          = "short"
	LIMIT                          = "limit"
	MARKET                         = "market"
	STOP_LOSS                      = "stop_loss"
	TAKE_PROFIT                    = "take_profit"
	STOP_LOSS_LIMIT                = "stop_loss_limit"
	TAKE_PROFIT_LIMIT              = "take_profit_limit"
	STOP_LIMIT                     = "stop_limit"
	STOP_MARKET                    = "stop_market"
	PING_LIMIT                     = "ping_limit"
	ACCOUNT_PREFIX                 = "account@"
	BALANCE_OPS_STATUS_PENDING     = "pending"
	BALANCE_OPS_STATUS_SUCCESS     = "success"
	BALANCE_OPS_STATUS_FAILED      = "failed"
	BALANCE_OPS_STATUS_TRANSFERING = "transferring"
	BALANCE_OPS_STATUS_UNKNOWN     = "unknown"
	BALANCE_OPS_TYPE_DEPOSIT       = "deposit"
	BALANCE_OPS_TYPE_WITHDRAW      = "withdraw"
	BALANCE_OPS_TYPE_STAKE         = "stake"
	BALANCE_OPS_TYPE_UNSTAKE       = "unstake"

	ERR_REFERRAL_PAYOUT_ID_DUPLICATE        = "ERR_REFERRAL_PAYOUT_ID_DUPLICATE"
	ERR_REFERRAL_PAYOUT_AMOUNT_NOT_POSITIVE = "ERR_REFERRAL_PAYOUT_AMOUNT_NOT_POSITIVE"

	GAMEASSETS_BLAST_LEADERBOARD_ROW_LIMIT     = 100
	GAMEASSETS_BLAST                           = "blast"
	GAMEASSETS_BLAST_LOAD_ASSETS_MAX_BATCH_LEN = 1000

	GAMEASSETS_BFX = "bfx"

)

var supportedProfileTypes = []string{PROFILE_TYPE_TRADER, PROFILE_TYPE_VAULT, PROFILE_TYPE_INSURANCE, PROFILE_TYPE_INSURANCE}

// order status
const (
	PLACED   = "placed"
	OPEN     = "open"
	CANCELED = "canceled"
	CLOSED   = "closed"
)

// airdrop status
const (
	AIRDROP_CLAIMING_STATUS = "claiming"
	AIRDROP_CLAIMED_STATUS  = "claimed"
)

// profile airdrop status
const (
	PROFILE_AIRDROP_INIT_STATUS     = "init"
	PROFILE_AIRDROP_ACTIVE_STATUS   = "active"
	PROFILE_AIRDROP_FINISHED_STATUS = "finished"
)

// ^uint(0) is platform dependent - will not use it
const MAX_PROFILE_ID = uint(4e9)

// for backward compability, will be removed later
const (
	API_SECRET_GEN_STATUS = "API_GEN"
)

const DEFAULT_INSTRUMENT_PRODUCT_TYPE = "perpetual"

// exchange ids
const (
	EXCHANGE_RBX = "rbx"
	EXCHANGE_BFX = "bfx"
)

var EXCHANGE_DEFAULT string = EXCHANGE_RBX

var SupportedExchangeIds = []string{EXCHANGE_RBX, EXCHANGE_BFX}
