-- Create the account table
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    account_number VARCHAR(20) UNIQUE NOT NULL,
    full_name VARCHAR(50) NOT NULL,
    id_number VARCHAR(16) UNIQUE NOT NULL,
    phone_number VARCHAR(15) UNIQUE NOT NULL,
    balance NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Function for auto-updating 'updated_at'
CREATE OR REPLACE FUNCTION update_accounts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update 'updated_at' on each update
CREATE TRIGGER update_accounts_trigger
BEFORE UPDATE ON accounts
FOR EACH ROW
EXECUTE PROCEDURE update_accounts_updated_at();

-- Add indexes for optimization (optional but recommended)
CREATE INDEX idx_accounts_account_number ON accounts(account_number);