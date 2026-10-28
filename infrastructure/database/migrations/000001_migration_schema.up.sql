CREATE TYPE account_role_enum AS ENUM ('seeker', 'recruiter');
CREATE TYPE workplace_enum AS ENUM ('remote', 'hybrid', 'on-site');
CREATE TYPE employment_enum AS ENUM ('full_time', 'part_time', 'freelance', 'contract', 'internship');
CREATE TYPE job_status_enum AS ENUM ('active', 'close', 'draft');
CREATE TYPE education_enum AS ENUM ('sd', 'smp', 'sma', 'smk', 'diploma', 'sarjana', 'magister', 'doktor');
CREATE TYPE submit_status_enum AS ENUM ('pending', 'reviewed', 'shortlisted', 'accepted', 'rejected');
CREATE TYPE payment_status_enum AS ENUM ('pending', 'success', 'settlement', 'cancel', 'expired');


CREATE TABLE accounts(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255), 
    role account_role_enum NOT NULL,
    phone_number VARCHAR(16), 
    auth_provider VARCHAR(50),
    refresh_token TEXT,
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE sectors(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sector_name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL UNIQUE REFERENCES accounts(id) ON DELETE CASCADE,
    sector_id UUID NOT NULL REFERENCES sectors(id) ON DELETE RESTRICT,
    company_name VARCHAR(255) NOT NULL,
    domicile VARCHAR(255) NOT NULL,
    web_url VARCHAR(255),
    address TEXT,
    owner_name VARCHAR(255) NOT NULL,
    description TEXT,
    avatar_url VARCHAR(500),
    average_rating REAL DEFAULT 0,
    total_reviews INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE applicants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL UNIQUE REFERENCES accounts(id) ON DELETE CASCADE,
    first_name VARCHAR(255), 
    last_name VARCHAR(255),  
    gender VARCHAR(50),
    birth_date DATE,
    domicile VARCHAR(255),
    avatar_url VARCHAR(500),
    resume_url VARCHAR(500),
    portfolio_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transactions(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    order_id VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    payment_type VARCHAR(100),
    status payment_status_enum NOT NULL DEFAULT 'pending',
    snap_token VARCHAR(255),
    snap_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, 
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE wallets(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    balance BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, 
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE job_category(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE jobs( 
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    job_category_id UUID NOT NULL REFERENCES job_category(id) ON DELETE CASCADE,
    position VARCHAR(255) NOT NULL,
    location VARCHAR(255) NOT NULL,
    workplace workplace_enum NOT NULL,
    employment employment_enum NOT NULL,
    salary_desc VARCHAR(255) NOT NULL,
    jobdesc TEXT NOT NULL,
    status job_status_enum DEFAULT 'active',
    slot INT,
    filled_slot INT NOT NULL DEFAULT 0, 
    coin_applyment INT NOT NULL DEFAULT 1,
    closing_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE essay_question(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    question_detail TEXT NOT NULL
);

CREATE TABLE submissions(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    applicant_id UUID NOT NULL REFERENCES applicants(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    web_portfolio_url VARCHAR(500),
    status submit_status_enum DEFAULT 'pending',
    internal_notes TEXT,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(applicant_id, job_id)
);

CREATE TABLE submissions_file(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submissions_id UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    file_type VARCHAR(50) NOT NULL,  
    file_name VARCHAR(255) NOT NULL, 
    file_size_bytes BIGINT,          
    file_url VARCHAR(500) NOT NULL,
    UNIQUE(submissions_id, file_url)
);

CREATE TABLE essay_answer(
    submissions_id UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    essay_question_id UUID NOT NULL REFERENCES essay_question(id) ON DELETE CASCADE,
    answer TEXT NOT null,
    PRIMARY KEY (submissions_id, essay_question_id)
);

CREATE TABLE educations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    applicant_id UUID NOT NULL REFERENCES applicants(id) ON DELETE CASCADE,
    school_name VARCHAR(255) NOT NULL,
    degree VARCHAR(100) NOT NULL, 
    field_of_study VARCHAR(255),
    start_date DATE NOT NULL,
    end_date DATE, 
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE experience (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    applicant_id UUID NOT NULL REFERENCES applicants(id) ON DELETE CASCADE,
    position VARCHAR(255) NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    current_job BOOLEAN DEFAULT false,
    jobdesc TEXT
);