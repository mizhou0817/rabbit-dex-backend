package api_client

const WS_PROD_URL = "wss://api.prod.rabbitx.io/ws"
const PROD_URL = "https://api.prod.rabbitx.io"

const WS_DEV_URL = "wss://api.dev.rabbitx.io/ws"
const DEV_URL = "https://api.dev.rabbitx.io"

const WS_TEST_URL = "wss://api.testnet.rabbitx.io/ws"
const TEST_URL = "https://api.testnet.rabbitx.io"
const LOCAL_URL = "http://localhost:8888"

const ONBOARDING_MESSAGE = "Welcome to RabbitX!\n\nClick to sign in and on-board your wallet for trading perpetuals.\n\nThis request will not trigger a blockchain transaction or cost any gas fees. This signature only proves you are the true owner of this wallet.\n\nBy signing this message you agree to the terms and conditions of the exchange."
const SIGNATURE_LIFETIME = 300

const PATH_ONBOARDING = "/onboarding"
const PATH_MARKETS = "/markets"
const PATH_ORDERS = "/orders"
const PATH_ORDERS_CANCEL_ALL = "/orders/cancel_all"

const PATH_ORDERS_LIST = "/orders/list"
const PATH_JWT = "/jwt"
const PATH_CANDLES = "/candles"
const PATH_ACCOUNT = "/account"
const PATH_ACCOUNT_LEVERAGE = "/account/leverage"
const PATH_ACCOUNT_VALIDATE = "/account/validate"
const PATH_POSITIONS = "/positions/list"
const PATH_DEPOSIT = "/balanceops/deposit"
const PATH_WITHDRAW = "/balanceops/withdraw"
const PATH_CANCEL_WITHDRAWAL = "/balanceops/cancel"
const PATH_CLAIM_WITHDRAWAL = "/balanceops/claim"
const PATH_ADMIN = "/admin"
const PATH_SUPER_ADMIN = "/super/admin"

const PATH_AIRDROP = "/airdrops"
const PATH_AIRDROP_CLAIM = "/airdrops/claim"
const PATH_AIRDROP_INIT = "/airdrops/init"
const PATH_AIRDROP_UPDATE = "/airdrops/update"

const PATH_SECRETS = "/secrets"
const PATH_STORAGE = "/storage"

const PATH_STATS_OI = "/stats/oi"
