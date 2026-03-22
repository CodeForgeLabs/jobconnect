# Payment Service Architecture

The **Payment Service** acts as the central financial orchestrator for the JobConnect platform. Its primary role is to bridge the gap between external Ethiopian payment gateways (**Telebirr** and **Chapa**) and the internal **Wallet Service**'s ledger system.

This service is designed to provide the same level of trust and security found in global platforms like Upwork, but tailored specifically for the Ethiopian market.

---

## 1. Core Responsibilities

1.  **Gateway Aggregation**: Normalizes the APIs of multiple local providers (Telebirr, Chapa) into a single internal interface.
2.  **External Sessions**: Manages the lifecycle of a deposit (e.g., generating a Telebirr checkout URL and verifying the callback).
3.  **Escrow Orchestration**: Coordinates with the `contract` and `wallet` services to "fund" milestones, ensuring the client's money is held securely until the freelancer delivers the work.
4.  **Fee Calculation**: Automatically calculates and deducts the platform service fee (e.g., 5-10%) before crediting a freelancer's wallet.
5.  **Disbursements (Withdrawals)**: Manages the process of moving earned funds from a freelancer's platform wallet to their personal bank/mobile money account.
6.  **Idempotency & Reconciliation**: Ensures that every transaction (especially via webhooks) is processed exactly once, even if the network fails.

---

## 2. Integration Points

The Payment Service is highly collaborative and sits at the center of the backend ecosystem:

| Service | Interaction Role | Typical Workflow |
| :--- | :--- | :--- |
| **Wallet Service** | **Ledger** | `payment` calls `wallet.Credit/Debit` to update internal balances after external verification. |
| **Contract Service** | **Business Logic** | `payment` notifies `contract` when a milestone is "FUNDED" so work can begin. |
| **Verification Service** | **Compliance (KYC)** | `payment` checks a user's verification status before allowing any withdrawal. |
| **Notification Service** | **UX** | `payment` triggers alerts (SMS/Email/In-app) when a deposit is confirmed or a payout fails. |

---

## 3. Core Workflows

### 3.1 Milestone Funding (Deposit & Escrow)
1.  **Request**: Client selects "Fund Milestone" on a contract.
2.  **Session Creation**: `payment` service calls Chapa/Telebirr to create a payment link and returns it to the client.
3.  **Client Pays**: The client completes the transaction on the provider's platform.
4.  **Webhook/Verification**: The provider sends a POST request (webhook) to the `payment` service confirming success.
5.  **Internal Sync**:
    -   `payment` calls `wallet.CreditWalletInternal` (Internal Credit).
    -   `payment` calls `wallet.PlaceHold` (Escrow/Hold for the milestone).
    -   `payment` calls `contract.UpdateMilestoneStatus` (to `FUNDED`).

### 3.2 Milestone Approval (The "Payday")
1.  **Approval**: Client approves the freelancer's work in the `contract` service.
2.  **Event**: `contract` service notifies `payment` (or `payment` polls for status).
3.  **Capture**: `payment` calls `wallet.CaptureHold` to move the money from "Held" to the freelancer's "Available" balance, minus the service fee.

### 3.3 Withdrawal (Payout)
1.  **Request**: Freelancer requests to withdraw 10,000 ETB to their bank account.
2.  **Verification Check**: `payment` calls `verification` to ensure the user is KYC-approved.
3.  **Debit Hold**: `payment` calls `wallet.DebitWalletInternal` to temporarily remove the funds from the available balance.
4.  **Bank Transfer**: `payment` calls the Chapa API to perform a "Transfer" (payout).
5.  **Completion**: Upon success, the `payment` service marks the transaction as "Success" and logs the external reference.

---

## 4. Payment State Machine

Each payment attempt follows this lifecycle:
-   `PENDING`: Session created, waiting for user to pay.
-   `COMPLETED`: Verified by webhook/gateway and internal wallets updated.
-   `FAILED`: Payment rejected by the provider or expired session.
-   `REFUNDED`: Money returned to the client (e.g., via dispute resolution).

---

## 5. Security & Idempotency

-   **JWT Authorization**: Only the account owner or an authorized Admin can initiate withdrawals.
-   **Webhook Signature Verification**: All incoming webhooks from Telebirr/Chapa MUST be cryptographically verified to prevent "fake payment" attacks.
-   **Idempotency Keys**: Every call to the `wallet` service includes a unique `idempotency_key` (usually `payment_attempt_id`) to prevent double-crediting if a webhook is retried.

---

## 6. Local Gateway Roadmap

-   [ ] **Chapa Integration**: Primary for Credit Cards, bank transfers, and payment aggregation.
-   [ ] **Telebirr Direct**: Direct integration with Ethio Telecom's API for low-fee mobile money.
