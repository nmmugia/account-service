-- Drop the triggers
DROP TRIGGER IF EXISTS update_account_balance ON cash_activities;

-- Drop the functions
DROP FUNCTION IF EXISTS update_account_balance();
DROP FUNCTION IF EXISTS get_latest_activity_id(INT);

-- Drop the cash_activity table
DROP TABLE IF EXISTS cash_activities;