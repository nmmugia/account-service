-- Drop the trigger
DROP TRIGGER IF EXISTS update_accounts_trigger ON accounts;

-- Drop the function
DROP FUNCTION IF EXISTS update_accounts_updated_at();

-- Drop the account table
DROP TABLE IF EXISTS accounts;