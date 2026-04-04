import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

interface LoginState {
	isLoggedIn: boolean;
	isClient: boolean;
	isFreelancer: boolean;
}

const initialState: LoginState = {
	isLoggedIn: false,
	isClient: false,
	isFreelancer: false,
};

const loginSlice = createSlice({
	name: "login",
	initialState,
	reducers: {
		setLoggedIn(state, action: PayloadAction<boolean>) {
			state.isLoggedIn = action.payload;
		},
		logIn(state) {
			state.isLoggedIn = true;
		},
		logOut(state) {
			state.isLoggedIn = false;
		},

		setIsClient(state, action: PayloadAction<boolean>) {
			state.isClient = action.payload;
		},
		setIsFreelancer(state, action: PayloadAction<boolean>) {
			state.isFreelancer = action.payload;
		},

	},
});

export const { setLoggedIn, logIn, logOut, setIsClient, setIsFreelancer } = loginSlice.actions;
export const selectIsLoggedIn = (state: { login: LoginState }) =>
	state.login.isLoggedIn;
export const selectIsClient = (state: { login: LoginState }) =>
	state.login.isClient;
export const selectIsFreelancer = (state: { login: LoginState }) =>
	state.login.isFreelancer;

export default loginSlice.reducer;
