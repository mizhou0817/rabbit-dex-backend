local decimal = require('decimal')

local default_min_initial_margin = decimal.new("0.05")
local default_forced_margin = decimal.new("0.03")
local default_liquidation_margin = decimal.new("0.02")
local default_adv_constant = decimal.new("5000000.0")
local role_btc = "market-btc"
local role_eth = "market-eth"
local role_sol = "market-sol"
local role_arb = "market-arb"
local role_doge = "market-doge"
local role_ldo = "market-ldo"
local role_sui = "market-sui"
local role_pepe = "market-pepe"
local role_bch = "market-bch"
local role_xrp = "market-xrp"
local role_wld = "market-wld"
local role_ton = "market-ton"
local role_stx = "market-stx"
local role_matic = "market-matic"
local role_trb = "market-trb"
local role_apt = "market-apt"
local role_inj = "market-inj"
local role_aave = "market-aave"
local role_link = "market-link"
local role_bnb = "market-bnb"
local role_rndr = "market-rndr"
local role_mkr = "market-mkr"
local role_rlb = "market-rlb"
local role_ordi = "market-ordi"
local role_stg = "market-stg"
local role_sats = "market-sats"
local role_tia = "market-tia"
local role_blur = "market-blur"
local role_jto = "market-jto"
local role_meme = "market-meme"
local role_sei = "market-sei"
local role_yes = "market-yes"
local role_wif = "market-wif"
local role_strk = "market-strk"
local role_shib = "market-shib"
local role_bome = "market-bome"
local role_slerf = "market-slerf"
local role_w = "market-w"
local role_ena = "market-ena"
local role_pac = "market-pac"
local role_maga = "market-maga"
local role_trump = "market-trump"
local role_mog = "market-mog"
local role_not = "market-not"
local role_mother = "market-mother"
local role_bonk = "market-bonk"
local role_taiko = "market-taiko"
local role_floki = "market-floki"

local role_test_market = "test_engine"

local params = {
    LONG = "long",
    SHORT = "short",
    JWT_PUBLIC =
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI0MDAwMDAwMDAwIiwiZXhwIjo2NTQ4NDg3NTY5fQ.o_qBZltZdDHBH3zHPQkcRhVBQCtejIuyq8V1yj5kYq8",

    LIQUIDATION = {
        FORCED_MARGIN = default_forced_margin,
        LIQUIDATION_MARGIN = default_liquidation_margin
    },

    PROFILE_TYPE = {
        TRADER = "trader",
        INSURANCE = "insurance",
        VAULT = "vault"
    },

    PROFILE_STATUS = {
        ACTIVE = "active",
        BLOCKED = "blocked",
        LIQUIDATING = "liquidating"
    },

    VAULT_STATUS = {
        ACTIVE = "active",
        SUSPENDED = "suspended",
    },

    MARKETS = {
        BTCUSDT = "BTC-USD",
        ETHUSDT = "ETH-USD",
        SOLUSDT = "SOL-USD",
        ARBUSDT = "ARB-USD",
        DOGEUSDT = "DOGE-USD",
        LDOUSDT = "LDO-USD",
        SUIUSDT = "SUI-USD",
        PEPE1000USDT = "PEPE1000-USD",
        BCHUSDT = "BCH-USD",
        XRPUSDT = "XRP-USD",
        WLDUSDT = "WLD-USD",
        TONUSDT = "TON-USD",
        STXUSDT = "STX-USD",
        MATICUSDT = "MATIC-USD",
        TRBUSDT = "TRB-USD",
        APTUSDT = "APT-USD",
        INJUSDT = "INJ-USD",
        AAVEUSDT = "AAVE-USD",
        LINKUSDT = "LINK-USD",
        BNBUSDT = "BNB-USD",
        RNDRUSDT = "RNDR-USD",
        MKRUSDT = "MKR-USD",
        RLBUSDT = "RLB-USD",
        ORDIUSDT = "ORDI-USD",
        STGUSDT = "STG-USD",
        SATS1000000USDT = "SATS1000000-USD",
        TIAUSDT = "TIA-USD",
        BLURUSDT = "BLUR-USD",
        JTOUSDT = "JTO-USD",
        MEMEUSDT = "MEME-USD",
        SEIUSDT = "SEI-USD",
        YESUSDT = "YES-USD",
        WIFUSDT = "WIF-USD",
        STRKUSDT = "STRK-USD",
        SHIB1000USDT = "SHIB1000-USD",
        BOMEUSDT = "BOME-USD",
        SLERFUSDT = "SLERF-USD",
        WUSDT = "W-USD",
        ENAUSDT = "ENA-USD",
        PACUSDT = "PAC-USD",
        MAGAUSDT = "MAGA-USD",
        TRUMPUSDT = "TRUMP-USD",
        MOG1000USDT = "MOG1000-USD",
        NOTUSDT = "NOT-USD",
        MOTHERUSDT = "MOTHER-USD",
        BONK1000USDT = "BONK1000-USD",
        TAIKOUSDT = "TAIKO-USD",
        FLOKI1000USDT = "FLOKI1000-USD",
    },

    MARKET_STATUS = {
        ACTIVE = "active",
        PAUSED = "paused",
    },

    ORDER_STATUS = {
        UNKNOWN      = "unknown",
        PROCESSING   = "processing",
        PLACED       = "placed",
        OPEN         = "open",
        CLOSED       = "closed",
        REJECTED     = "rejected",
        CANCELED     = "canceled",
        CANCELING    = "canceling",
        AMENDING     = "amending",
        CANCELINGALL = "cancelingall"
    },

    AIRDROP_CLAIM_STATUS = {
        CLAIMING = "claiming",
        CLAIMED  = "claimed"
    },

    PROFILE_AIRDROP_STATUS = {
        INIT = "init",
        ACTIVE = "active",
        FINISHED = "finished"
    },

    ORDER_TYPE = {
        MARKET = "market",
        LIMIT = "limit",
        STOP_LOSS = "stop_loss",
        TAKE_PROFIT = "take_profit",
        STOP_LOSS_LIMIT = "stop_loss_limit",
        TAKE_PROFIT_LIMIT = "take_profit_limit",
        STOP_LIMIT = "stop_limit",
        STOP_MARKET = "stop_market",
        PING_LIMIT = "ping_limit",
    },

    TIME_IN_FORCE = {
        GTC = "good_till_cancel",
        IOC = "immediate_or_cancel",
        FOK = "fill_or_kill",
        POST_ONLY = "post_only",
    },

    ORDER_ACTION = {
        CREATE = "create",
        AMEND = "amend",
        CANCEL = "cancel",
        CANCELALL = "cancelall",
        LIQUIDATE = "liquidate",
        EXECUTE = "execute",
    },

    LIQUIDATE_KIND = {
        APLACESELLORDERS = 0,
        AINSTAKEOVER = 1,
        AINSCLAWBACK = 2
    },

    BALANCE_STATUS = {
        PENDING    = "pending",
        REQUESTED  = "requested",
        SUCCESS    = "success",    -- when it's claimed or deposited or other
        PROCESSING = "processing", -- when we have a txhash but no event log confirmation
        UNKNOWN    = "unknown",
        CLAIMABLE  = "claimable",  -- withdrawal once 6 hours passed
        CLAIMING   = "claiming",   -- once you requested the signature
        CANCELED   = "canceled",   -- possible only for pending or claimable
    },

    BALANCE_TYPE = {
        DEPOSIT              = "deposit",
        CREDIT               = "credit",
        WITHDRAWAL           = "withdrawal",
        WITHDRAW_CREDIT      = "withdraw_credit",
        FUNDING              = "funding",
        PNL                  = "pnl",
        FEE                  = "fee",
        WITHDRAW_FEE         = "withdraw_fee",
        REFERRAL_PAYOUT      = "referral_payout",
        YIELD_PAYOUT         = "yield_payout",

        -- balancing operations...

        -- either one of these:
        STAKE                = "stake",              -- staker balance unaffected
        STAKE_FROM_BALANCE   = "stake_from_balance", -- staker balance goes down

        -- is balanced by one of these:
        VAULT_STAKE          = "vault_stake",  -- "vs_" prefix, vault balance up

        STAKE_SHARES         = "stake_shares", -- "ss_" prefix, profile=staker, amount=shares

        -- each of these:
        UNSTAKE_SHARES       = "unstake_shares", -- profile=staker, amount=shares

        -- is balanced by one of these (for display, neither affects balances):
        VAULT_UNSTAKE_SHARES = "vault_unstake_shares", -- "vus_" prfx, prof=vault, amnt=shares


        -- each pair of these:
        UNSTAKE_VALUE       = "unstake_value", -- "uv_" prefix, staker balance up
        UNSTAKE_FEE         = "unstake_fee",   -- "uf_" prefix, treasurer balance up

        -- is balanced by one of these:
        VAULT_UNSTAKE_VALUE = "vault_unstake_value", -- "vuv_" prefix, vault balance down

    },

    NOTIF_TYPE = {
        NOTIF_ERROR = "error",
        NOTIF_WARNING = "warning",
        NOTIF_SUCCESS = "success",
        NOTIF_INFO = "info",
        NOTIF_DEBUG = "debug"
    },

    REVERT_STATUS = {
        TX_UPLOADED = "uploaded",
        TX_PROCESSED = "processed",
        TX_FAILED = "failed",
        TX_EXECUTED = "executed",
    },

    MAX_POSITION_DEPENDENT_ORDERS = 1,
    MAX_CONDITIONAL_ORDERS = 10,

    LIMIT_BUY_RATIO = decimal.new("1.03"),
    LIMIT_SELL_RATIO = decimal.new("0.97"),

    WAIT_ROLE_RETRY_INTERVAL = 3,
    MARKET_UPDATE_RETRY_INTERVAL = 1,

    MAX_SECRETS_PER_ACCOUNT = 100,
}

local errors = {
    NOT_FOUND = "NOT_FOUND"
}

local markets = {
    [params.MARKETS.BTCUSDT] = {
        id = params.MARKETS.BTCUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("1.0"),     -- min_tick
        min_order = decimal.new("0.0001"), -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },
    [params.MARKETS.ETHUSDT] = {
        id = params.MARKETS.ETHUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.1"),    -- min_tick
        min_order = decimal.new("0.001"), -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },
    [params.MARKETS.SOLUSDT] = {
        id = params.MARKETS.SOLUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("0.01"),  -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.ARBUSDT] = {
        id = params.MARKETS.ARBUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("1.0"),   -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.DOGEUSDT] = {
        id = params.MARKETS.DOGEUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.00001"), -- min_tick
        min_order = decimal.new("10.0"),   -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.LDOUSDT] = {
        id = params.MARKETS.LDOUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.001"), -- min_tick
        min_order = decimal.new("1.0"),  -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.SUIUSDT] = {
        id = params.MARKETS.SUIUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("1.0"),   -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.PEPE1000USDT] = {
        id = params.MARKETS.PEPE1000USDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.000001"), -- min_tick
        min_order = decimal.new("1000.0"),  -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.BCHUSDT] = {
        id = params.MARKETS.BCHUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.01"),  -- min_tick
        min_order = decimal.new("0.01"), -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.XRPUSDT] = {
        id = params.MARKETS.XRPUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("10.0"),  -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.WLDUSDT] = {
        id = params.MARKETS.WLDUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("1.0"),   -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.TONUSDT] = {
        id = params.MARKETS.TONUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("1.0"),   -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.STXUSDT] = {
        id = params.MARKETS.STXUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("1.0"),   -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.MATICUSDT] = {
        id = params.MARKETS.MATICUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("1.0"),   -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.TRBUSDT] = {
        id = params.MARKETS.TRBUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.01"),  -- min_tick
        min_order = decimal.new("0.01"), -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.APTUSDT] = {
        id = params.MARKETS.APTUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.001"), -- min_tick
        min_order = decimal.new("0.1"),  -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.INJUSDT] = {
        id = params.MARKETS.INJUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.001"), -- min_tick
        min_order = decimal.new("0.1"),  -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.AAVEUSDT] = {
        id = params.MARKETS.AAVEUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.01"),  -- min_tick
        min_order = decimal.new("0.01"), -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.LINKUSDT] = {
        id = params.MARKETS.LINKUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.001"), -- min_tick
        min_order = decimal.new("0.1"),  -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.BNBUSDT] = {
        id = params.MARKETS.BNBUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.01"),  -- min_tick
        min_order = decimal.new("0.01"), -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.RNDRUSDT] = {
        id = params.MARKETS.RNDRUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("1.0"),   -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.MKRUSDT] = {
        id = params.MARKETS.MKRUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.1"),    -- min_tick
        min_order = decimal.new("0.001"), -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.RLBUSDT] = {
        id = params.MARKETS.RLBUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("10.0"),  -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.ORDIUSDT] = {
        id = params.MARKETS.ORDIUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.001"), -- min_tick
        min_order = decimal.new("0.1"),  -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.STGUSDT] = {
        id = params.MARKETS.STGUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("10"),    -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.SATS1000000USDT] = {
        id = params.MARKETS.SATS1000000USDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("10"),    -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.TIAUSDT] = {
        id = params.MARKETS.TIAUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.001"), -- min_tick
        min_order = decimal.new("0.1"),  -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.BLURUSDT] = {
        id = params.MARKETS.BLURUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("1"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.JTOUSDT] = {
        id = params.MARKETS.JTOUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("1"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.MEMEUSDT] = {
        id = params.MARKETS.MEMEUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.00001"), -- min_tick
        min_order = decimal.new("10"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.SEIUSDT] = {
        id = params.MARKETS.SEIUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("1"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.YESUSDT] = {
        id = params.MARKETS.YESUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("0.1"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.WIFUSDT] = {
        id = params.MARKETS.WIFUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("1"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.STRKUSDT] = {
        id = params.MARKETS.STRKUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("1"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.SHIB1000USDT] = {
        id = params.MARKETS.SHIB1000USDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.000001"), -- min_tick
        min_order = decimal.new("100"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.BOMEUSDT] = {
        id = params.MARKETS.BOMEUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.000001"), -- min_tick
        min_order = decimal.new("100"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.SLERFUSDT] = {
        id = params.MARKETS.SLERFUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.00001"), -- min_tick
        min_order = decimal.new("1"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.WUSDT] = {
        id = params.MARKETS.WUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("1"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.ENAUSDT] = {
        id = params.MARKETS.ENAUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("1"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.PACUSDT] = {
        id = params.MARKETS.PACUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.00001"), -- min_tick
        min_order = decimal.new("100"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.MAGAUSDT] = {
        id = params.MARKETS.MAGAUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0000001"), -- min_tick
        min_order = decimal.new("1000"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.TRUMPUSDT] = {
        id = params.MARKETS.TRUMPUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.001"), -- min_tick
        min_order = decimal.new("0.1"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.MOG1000USDT] = {
        id = params.MARKETS.MOG1000USDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.000001"), -- min_tick
        min_order = decimal.new("1000"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.NOTUSDT] = {
        id = params.MARKETS.NOTUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.000001"), -- min_tick
        min_order = decimal.new("100"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.MOTHERUSDT] = {
        id = params.MARKETS.MOTHERUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.0001"), -- min_tick
        min_order = decimal.new("10"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.BONK1000USDT] = {
        id = params.MARKETS.BONK1000USDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.000001"), -- min_tick
        min_order = decimal.new("100"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.TAIKOUSDT] = {
        id = params.MARKETS.TAIKOUSDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.001"), -- min_tick
        min_order = decimal.new("1"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },

    [params.MARKETS.FLOKI1000USDT] = {
        id = params.MARKETS.FLOKI1000USDT,
        status = params.MARKET_STATUS.ACTIVE,

        min_initial_margin = default_min_initial_margin,
        forced_margin = default_forced_margin,
        liquidation_margin = default_liquidation_margin,

        min_tick = decimal.new("0.00001"), -- min_tick
        min_order = decimal.new("10"),     -- min_order
        adv_constant = default_adv_constant,

        inv3_buffer_size = decimal.new(0),      -- inv3_buffer_size
        adv_ratio = decimal.new("0.05"),        -- adv_ratio
        limit_buy_ratio = decimal.new("1.05"),  -- limit_buy_ratio
        limit_sell_ratio = decimal.new("0.95"), -- limit_sell_ratio
        sltp_limit_buy_ratio = decimal.new("1.03"),
        sltp_limit_sell_ratio = decimal.new("0.97"),
    },
}


local sys = {
    ID_TO_ROLES = {
        ["TEST-MARKET"] = role_test_market,
        [params.MARKETS.BTCUSDT] = role_btc,
        [params.MARKETS.ETHUSDT] = role_eth,
        [params.MARKETS.SOLUSDT] = role_sol,
        [params.MARKETS.ARBUSDT] = role_arb,
        [params.MARKETS.DOGEUSDT] = role_doge,
        [params.MARKETS.LDOUSDT] = role_ldo,
        [params.MARKETS.SUIUSDT] = role_sui,
        [params.MARKETS.PEPE1000USDT] = role_pepe,
        [params.MARKETS.BCHUSDT] = role_bch,
        [params.MARKETS.XRPUSDT] = role_xrp,
        [params.MARKETS.WLDUSDT] = role_wld,
        [params.MARKETS.TONUSDT] = role_ton,
        [params.MARKETS.STXUSDT] = role_stx,
        [params.MARKETS.MATICUSDT] = role_matic,
        [params.MARKETS.TRBUSDT] = role_trb,
        [params.MARKETS.APTUSDT] = role_apt,
        [params.MARKETS.INJUSDT] = role_inj,
        [params.MARKETS.AAVEUSDT] = role_aave,
        [params.MARKETS.LINKUSDT] = role_link,
        [params.MARKETS.BNBUSDT] = role_bnb,
        [params.MARKETS.RNDRUSDT] = role_rndr,
        [params.MARKETS.MKRUSDT] = role_mkr,
        [params.MARKETS.RLBUSDT] = role_rlb,
        [params.MARKETS.ORDIUSDT] = role_ordi,
        [params.MARKETS.STGUSDT] = role_stg,
        [params.MARKETS.SATS1000000USDT] = role_sats,
        [params.MARKETS.TIAUSDT] = role_tia,
        [params.MARKETS.BLURUSDT] = role_blur,
        [params.MARKETS.JTOUSDT] = role_jto,
        [params.MARKETS.MEMEUSDT] = role_meme,
        [params.MARKETS.SEIUSDT] = role_sei,
        [params.MARKETS.YESUSDT] = role_yes,
        [params.MARKETS.WIFUSDT] = role_wif,
        [params.MARKETS.STRKUSDT] = role_strk,
        [params.MARKETS.SHIB1000USDT] = role_shib,
        [params.MARKETS.BOMEUSDT] = role_bome,
        [params.MARKETS.SLERFUSDT] = role_slerf,
        [params.MARKETS.WUSDT] = role_w,
        [params.MARKETS.ENAUSDT] = role_ena,
        [params.MARKETS.PACUSDT] = role_pac,
        [params.MARKETS.MAGAUSDT] = role_maga,
        [params.MARKETS.TRUMPUSDT] = role_trump,
        [params.MARKETS.MOG1000USDT] = role_mog,
        [params.MARKETS.NOTUSDT] = role_not,
        [params.MARKETS.MOTHERUSDT] = role_mother,
        [params.MARKETS.BONK1000USDT] = role_bonk,
        [params.MARKETS.TAIKOUSDT] = role_taiko,
        [params.MARKETS.FLOKI1000USDT] = role_floki,
    },
    ROLES = {
        btc = role_btc,
        eth = role_eth,
        sol = role_sol,
        arb = role_arb,
        doge = role_doge,
        ldo = role_ldo,
        sui = role_sui,
        pepe = role_pepe,
        bch = role_bch,
        xrp = role_xrp,
        wld = role_wld,
        ton = role_ton,
        stx = role_stx,
        matic = role_matic,
        trb = role_trb,
        apt = role_apt,
        inj = role_inj,
        aave = role_aave,
        link = role_link,
        bnb = role_bnb,
        rndr = role_rndr,
        mkr = role_mkr,
        rlb = role_rlb,
        ordi = role_ordi,
        stg = role_stg,
        sats = role_sats,
        tia = role_tia,
        blur = role_blur,
        jto = role_jto,
        meme = role_meme,
        sei = role_sei,
        yes = role_yes,
        wif = role_wif,
        strk = role_strk,
        shib = role_shib,
        bome = role_bome,
        slerf = role_slerf,
        w = role_w,
        ena = role_ena,
        pac = role_pac,
        maga = role_maga,
        trump = role_trump,
        mog = role_mog,
        _not = role_not,
        mother = role_mother,
        bonk = role_bonk,
        taiko = role_taiko,
        floki = role_floki,
        test_market = role_test_market
    },
    ID_TO_MARKET_QUEUE = {
        ["TEST-MARKET"] = { name = "queue-test-market", buffer_size = 1000 },
        [params.MARKETS.BTCUSDT] = { name = "queue-btc", buffer_size = 1000 },
        [params.MARKETS.ETHUSDT] = { name = "queue-eth", buffer_size = 1000 },
        [params.MARKETS.SOLUSDT] = { name = "queue-sol", buffer_size = 1000 },
        [params.MARKETS.ARBUSDT] = { name = "queue-arb", buffer_size = 1000 },
        [params.MARKETS.DOGEUSDT] = { name = "queue-doge", buffer_size = 1000 },
        [params.MARKETS.LDOUSDT] = { name = "queue-ldo", buffer_size = 1000 },
        [params.MARKETS.SUIUSDT] = { name = "queue-sui", buffer_size = 1000 },
        [params.MARKETS.PEPE1000USDT] = { name = "queue-pepe", buffer_size = 1000 },
        [params.MARKETS.BCHUSDT] = { name = "queue-bch", buffer_size = 1000 },
        [params.MARKETS.XRPUSDT] = { name = "queue-xrp", buffer_size = 1000 },
        [params.MARKETS.WLDUSDT] = { name = "queue-wld", buffer_size = 1000 },
        [params.MARKETS.TONUSDT] = { name = "queue-ton", buffer_size = 1000 },
        [params.MARKETS.STXUSDT] = { name = "queue-stx", buffer_size = 1000 },
        [params.MARKETS.MATICUSDT] = { name = "queue-matic", buffer_size = 1000 },
        [params.MARKETS.TRBUSDT] = { name = "queue-trb", buffer_size = 1000 },
        [params.MARKETS.APTUSDT] = { name = "queue-apt", buffer_size = 1000 },
        [params.MARKETS.INJUSDT] = { name = "queue-inj", buffer_size = 1000 },
        [params.MARKETS.AAVEUSDT] = { name = "queue-aave", buffer_size = 1000 },
        [params.MARKETS.LINKUSDT] = { name = "queue-link", buffer_size = 1000 },
        [params.MARKETS.BNBUSDT] = { name = "queue-bnb", buffer_size = 1000 },
        [params.MARKETS.RNDRUSDT] = { name = "queue-rndr", buffer_size = 1000 },
        [params.MARKETS.MKRUSDT] = { name = "queue-mkr", buffer_size = 1000 },
        [params.MARKETS.RLBUSDT] = { name = "queue-rlb", buffer_size = 1000 },
        [params.MARKETS.ORDIUSDT] = { name = "queue-ordi", buffer_size = 1000 },
        [params.MARKETS.STGUSDT] = { name = "queue-stg", buffer_size = 1000 },
        [params.MARKETS.SATS1000000USDT] = { name = "queue-sats", buffer_size = 1000 },
        [params.MARKETS.TIAUSDT] = { name = "queue-tia", buffer_size = 1000 },
        [params.MARKETS.BLURUSDT] = { name = "queue-blur", buffer_size = 1000 },
        [params.MARKETS.JTOUSDT] = { name = "queue-jto", buffer_size = 1000 },
        [params.MARKETS.MEMEUSDT] = { name = "queue-meme", buffer_size = 1000 },
        [params.MARKETS.SEIUSDT] = { name = "queue-sei", buffer_size = 1000 },
        [params.MARKETS.YESUSDT] = { name = "queue-yes", buffer_size = 1000 },
        [params.MARKETS.WIFUSDT] = { name = "queue-wif", buffer_size = 1000 },
        [params.MARKETS.STRKUSDT] = { name = "queue-strk", buffer_size = 1000 },
        [params.MARKETS.SHIB1000USDT] = { name = "queue-shib", buffer_size = 1000 },
        [params.MARKETS.BOMEUSDT] = { name = "queue-bome", buffer_size = 1000 },
        [params.MARKETS.SLERFUSDT] = { name = "queue-slerf", buffer_size = 1000 },
        [params.MARKETS.WUSDT] = { name = "queue-w", buffer_size = 1000 },
        [params.MARKETS.ENAUSDT] = { name = "queue-ena", buffer_size = 1000 },
        [params.MARKETS.PACUSDT] = { name = "queue-pac", buffer_size = 1000 },
        [params.MARKETS.MAGAUSDT] = { name = "queue-maga", buffer_size = 1000 },
        [params.MARKETS.TRUMPUSDT] = { name = "queue-trump", buffer_size = 1000 },
        [params.MARKETS.MOG1000USDT] = { name = "queue-mog", buffer_size = 1000 },
        [params.MARKETS.NOTUSDT] = { name = "queue-not", buffer_size = 1000 },
        [params.MARKETS.MOTHERUSDT] = { name = "queue-mother", buffer_size = 1000 },
        [params.MARKETS.BONK1000USDT] = { name = "queue-bonk", buffer_size = 1000 },
        [params.MARKETS.TAIKOUSDT] = { name = "queue-taiko", buffer_size = 1000 },
        [params.MARKETS.FLOKI1000USDT] = { name = "queue-floki", buffer_size = 1000 },
    },
    ID_TO_LIQUIDATION_QUEUE = {
        ["TEST-MARKET"] = { name = "liq-test-market", buffer_size = 1000 },
        [params.MARKETS.BTCUSDT] = { name = "liq-queue-btc", buffer_size = 1000 },
        [params.MARKETS.ETHUSDT] = { name = "liq-queue-eth", buffer_size = 1000 },
        [params.MARKETS.SOLUSDT] = { name = "liq-queue-sol", buffer_size = 1000 },
        [params.MARKETS.ARBUSDT] = { name = "liq-queue-arb", buffer_size = 1000 },
        [params.MARKETS.DOGEUSDT] = { name = "liq-queue-doge", buffer_size = 1000 },
        [params.MARKETS.LDOUSDT] = { name = "liq-queue-ldo", buffer_size = 1000 },
        [params.MARKETS.SUIUSDT] = { name = "liq-queue-sui", buffer_size = 1000 },
        [params.MARKETS.PEPE1000USDT] = { name = "liq-queue-pepe", buffer_size = 1000 },
        [params.MARKETS.BCHUSDT] = { name = "liq-queue-bch", buffer_size = 1000 },
        [params.MARKETS.XRPUSDT] = { name = "liq-queue-xrp", buffer_size = 1000 },
        [params.MARKETS.WLDUSDT] = { name = "liq-queue-wld", buffer_size = 1000 },
        [params.MARKETS.TONUSDT] = { name = "liq-queue-ton", buffer_size = 1000 },
        [params.MARKETS.STXUSDT] = { name = "liq-queue-stx", buffer_size = 1000 },
        [params.MARKETS.MATICUSDT] = { name = "liq-queue-matic", buffer_size = 1000 },
        [params.MARKETS.TRBUSDT] = { name = "liq-queue-trb", buffer_size = 1000 },
        [params.MARKETS.APTUSDT] = { name = "liq-queue-apt", buffer_size = 1000 },
        [params.MARKETS.INJUSDT] = { name = "liq-queue-inj", buffer_size = 1000 },
        [params.MARKETS.AAVEUSDT] = { name = "liq-queue-aave", buffer_size = 1000 },
        [params.MARKETS.LINKUSDT] = { name = "liq-queue-link", buffer_size = 1000 },
        [params.MARKETS.BNBUSDT] = { name = "liq-queue-bnb", buffer_size = 1000 },
        [params.MARKETS.RNDRUSDT] = { name = "liq-queue-rndr", buffer_size = 1000 },
        [params.MARKETS.MKRUSDT] = { name = "liq-queue-mkr", buffer_size = 1000 },
        [params.MARKETS.RLBUSDT] = { name = "liq-queue-rlb", buffer_size = 1000 },
        [params.MARKETS.ORDIUSDT] = { name = "liq-queue-ordi", buffer_size = 1000 },
        [params.MARKETS.STGUSDT] = { name = "liq-queue-stg", buffer_size = 1000 },
        [params.MARKETS.SATS1000000USDT] = { name = "liq-queue-sats", buffer_size = 1000 },
        [params.MARKETS.TIAUSDT] = { name = "liq-queue-tia", buffer_size = 1000 },
        [params.MARKETS.BLURUSDT] = { name = "liq-queue-blur", buffer_size = 1000 },
        [params.MARKETS.JTOUSDT] = { name = "liq-queue-jto", buffer_size = 1000 },
        [params.MARKETS.MEMEUSDT] = { name = "liq-queue-meme", buffer_size = 1000 },
        [params.MARKETS.SEIUSDT] = { name = "liq-queue-sei", buffer_size = 1000 },
        [params.MARKETS.YESUSDT] = { name = "liq-queue-yes", buffer_size = 1000 },
        [params.MARKETS.WIFUSDT] = { name = "liq-queue-wif", buffer_size = 1000 },
        [params.MARKETS.STRKUSDT] = { name = "liq-queue-strk", buffer_size = 1000 },
        [params.MARKETS.SHIB1000USDT] = { name = "liq-queue-shib", buffer_size = 1000 },
        [params.MARKETS.BOMEUSDT] = { name = "liq-queue-bome", buffer_size = 1000 },
        [params.MARKETS.SLERFUSDT] = { name = "liq-queue-slerf", buffer_size = 1000 },
        [params.MARKETS.WUSDT] = { name = "liq-queue-w", buffer_size = 1000 },
        [params.MARKETS.ENAUSDT] = { name = "liq-queue-ena", buffer_size = 1000 },
        [params.MARKETS.PACUSDT] = { name = "liq-queue-pac", buffer_size = 1000 },
        [params.MARKETS.MAGAUSDT] = { name = "liq-queue-maga", buffer_size = 1000 },
        [params.MARKETS.TRUMPUSDT] = { name = "liq-queue-trump", buffer_size = 1000 },
        [params.MARKETS.NOTUSDT] = { name = "liq-queue-not", buffer_size = 1000 },
        [params.MARKETS.MOTHERUSDT] = { name = "liq-queue-mother", buffer_size = 1000 },
        [params.MARKETS.BONK1000USDT] = { name = "liq-queue-bonk", buffer_size = 1000 },
        [params.MARKETS.TAIKOUSDT] = { name = "liq-queue-taiko", buffer_size = 1000 },
        [params.MARKETS.FLOKI1000USDT] = { name = "liq-queue-floki", buffer_size = 1000 },
    },

    QUEUE_TYPE = {
        MARKET = "market",
        LIQUIDATION = "liq"
    },

    -- all time in microseconds
    QUEUE_LIMITS = {
        PERIOD = 1e6,
        LIMIT = 20,
        TOTAL_SIZE = 50000
    },

    EVENTS = {
        PROFILE_UPDATE = "profile_update"
    },

    MODE = "async"
}

return {
    markets = markets,
    sys = sys,
    params = params
}
