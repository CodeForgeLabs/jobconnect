DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'auth') THEN
        CREATE ROLE auth LOGIN PASSWORD 'auth';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'user') THEN
        CREATE ROLE "user" LOGIN PASSWORD 'user';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'job') THEN
        CREATE ROLE job LOGIN PASSWORD 'job';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'proposal') THEN
        CREATE ROLE proposal LOGIN PASSWORD 'proposal';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'contract') THEN
        CREATE ROLE contract LOGIN PASSWORD 'contract';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'wallet') THEN
        CREATE ROLE wallet LOGIN PASSWORD 'wallet';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'chat') THEN
        CREATE ROLE chat LOGIN PASSWORD 'chat';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'connects') THEN
        CREATE ROLE connects LOGIN PASSWORD 'connects';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'verification') THEN
        CREATE ROLE verification LOGIN PASSWORD 'verification';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'payment') THEN
        CREATE ROLE payment LOGIN PASSWORD 'payment';
    END IF;
END
$$;

SELECT 'CREATE DATABASE jobconnect_auth OWNER auth'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'jobconnect_auth')\gexec

SELECT 'CREATE DATABASE jobconnect_user OWNER "user"'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'jobconnect_user')\gexec

SELECT 'CREATE DATABASE jobconnect_job OWNER job'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'jobconnect_job')\gexec

SELECT 'CREATE DATABASE jobconnect_proposal OWNER proposal'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'jobconnect_proposal')\gexec

SELECT 'CREATE DATABASE jobconnect_contract OWNER contract'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'jobconnect_contract')\gexec

SELECT 'CREATE DATABASE jobconnect_wallet OWNER wallet'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'jobconnect_wallet')\gexec

SELECT 'CREATE DATABASE jobconnect_chat OWNER chat'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'jobconnect_chat')\gexec

SELECT 'CREATE DATABASE jobconnect_connects OWNER connects'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'jobconnect_connects')\gexec

SELECT 'CREATE DATABASE jobconnect_verification OWNER verification'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'jobconnect_verification')\gexec

SELECT 'CREATE DATABASE jobconnect_payment OWNER payment'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'jobconnect_payment')\gexec
