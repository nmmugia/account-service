-- Create the cash_activity table
CREATE TABLE cash_activities (
    id SERIAL PRIMARY KEY,
    account_id bigint NOT NULL,
    reference_id INT, -- New column for chained transactions
    type VARCHAR(10) NOT NULL CHECK (type IN ('debit', 'credit')),
    nominal NUMERIC(15, 2) NOT NULL,
    balance_before NUMERIC(15, 2) NOT NULL,
    balance_after NUMERIC(15, 2) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (account_id) REFERENCES accounts(id)
);

-- Add indexes for optimization
CREATE INDEX idx_cash_activities_account_id ON cash_activities(account_id);
CREATE INDEX idx_cash_activities_reference_id ON cash_activities(reference_id); -- Index on reference_id
CREATE INDEX idx_cash_activities_created_at ON cash_activities(created_at);

-- Function to get the latest activity_id for an account
CREATE OR REPLACE FUNCTION get_latest_activity_id(p_account_id INT)
RETURNS INT AS $$
DECLARE
    v_latest_activity_id INT;
BEGIN
    SELECT id INTO v_latest_activity_id
    FROM cash_activities
    WHERE account_id = p_account_id
    ORDER BY created_at DESC
    LIMIT 1;
    
    RETURN v_latest_activity_id;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update balance in account table after cash activity
CREATE OR REPLACE FUNCTION update_account_balance()
RETURNS TRIGGER AS $$
DECLARE
    v_latest_activity_id INT;
BEGIN
  -- Get the latest activity_id for the account
  v_latest_activity_id := get_latest_activity_id(NEW.account_id);

  IF NEW.type = 'credit' THEN
    UPDATE accounts SET balance = balance + NEW.nominal WHERE id = NEW.account_id;
  ELSE
    UPDATE accounts SET balance = balance - NEW.nominal WHERE id = NEW.account_id;
  END IF;

  -- Update the reference_id with the latest activity_id
  NEW.reference_id := v_latest_activity_id;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;