import { baseApi } from "./baseapi";

export interface Wallet {
	ID: number;
	UserID: number;
	BalanceMinor: number;
	Currency: string;
	CreatedAt: string;
	UpdatedAt: string;
}

export interface WalletTransaction {
	ID: number;
	WalletID: number;
	TxRef: string;
	Type: string;
	Status: string;
	AmountMinor: number;
	Description: string;
	Provider: string;
	ExternalRef: string;
	CreatedAt: string;
	UpdatedAt: string;
}

export interface WalletBalanceResponse {
	wallet: Wallet;
}

export interface WalletTransactionsResponse {
	transactions: WalletTransaction[];
}

export interface BuyConnectInput {
	amount: number;
}

export interface CreateWalletTransactionInput {
	amountMinor: number;
	description: string;
	email: string;
	phone: string;
	userID: number;
}

export interface BuyConnectResponse {
	message: string;
}

export interface CreateWalletTransactionResponse {
	payment_url: string;
}

export const walletApi = baseApi.injectEndpoints({
	endpoints: (builder) => ({
		getWalletBalance: builder.query<Wallet, void>({
			query: () => "/wallet/balance",
			transformResponse: (response: WalletBalanceResponse) => response.wallet,
		}),

		buyConnect: builder.mutation<BuyConnectResponse, BuyConnectInput>({
			query: (body) => ({
				url: "/wallet/buy-connect",
				method: "POST",
				body,
			}),
		}),

		createWalletTransaction: builder.mutation<
			CreateWalletTransactionResponse,
			CreateWalletTransactionInput
		>({
			query: (body) => ({
				url: "/wallet/transaction",
				method: "POST",
				body,
			}),
		}),

		getWalletTransactions: builder.query<WalletTransaction[], void>({
			query: () => "/wallet/transactions",
			transformResponse: (response: WalletTransactionsResponse) =>
				response.transactions,
		}),
	}),
});

export const {
	useGetWalletBalanceQuery,
	useBuyConnectMutation,
	useCreateWalletTransactionMutation,
	useGetWalletTransactionsQuery,
} = walletApi;
