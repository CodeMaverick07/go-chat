package store

import (
	"database/sql"
	"errors"
	"fmt"
	Email "go-chat/internals/email"
	"go-chat/internals/utils"
	"time"
)


type OTPPurpose string
const (
	OTPPurposeVerify string = "verify"
	OTPPurposeLogin string = "login"
)
type OTP struct {
	ID string `json:"id"`
	Email string `json:"email"`
	CodeHash string `json:"code_hash"`
	Purpose OTPPurpose `json:"purpose"`
	ExpiresAt time.Time `json:"expires_at"`
	Used bool `json:"used"`
	Attempts string `json:"attempts"`
	MaxAttempts string `json:"max_attempts"`
	CreatedAt string `json:"created_at"`
}

type PostgresOTPStore struct {
	DB *sql.DB
	EmailSender *Email.Sender

}

func NewOTPStore(DB *sql.DB,EmailSender *Email.Sender) *PostgresOTPStore {
	return &PostgresOTPStore{
		DB: DB,
		EmailSender: EmailSender,
	}
}

type OTPstore interface{
	SendOTP(username string,email string,purpose OTPPurpose) error 
	VerifyOTP(email string,code string,purpose OTPPurpose)(*OTP,error)
}

func (p *PostgresOTPStore) SendOTP(username string,email string,purpose OTPPurpose) error {
	otp,err := utils.GenerateOTP()
	if err != nil {
		fmt.Println("not able to generate otp",err,otp)
		return err
	}
		otpHash,err := utils.Hash(otp.Code)
	if err != nil {
		fmt.Println("not able to hash the otp",err)
		return err
	}
	query := `
	INSERT INTO otp_codes (email,code_hash,purpose,expires_at)
	VALUES ($1,$2,$3,$4)
	RETURNING id
	`
	var id string
	err = p.DB.QueryRow(query,email,otpHash,purpose,otp.ExpiresAt).Scan(&id)
	fmt.Println("id:",id)
	if err != nil {
		fmt.Println("not able run send otp query",err)
		return err
	}
	
	subject,htmlBody,_:=Email.OTPVerificationTemplate(username,otp.Code)
	err = p.EmailSender.Send(email,subject,htmlBody)
	if err != nil {
		fmt.Println("not able to send otp",err)
		return err
	}


	return nil
}
func (p *PostgresOTPStore) VerifyOTP(
	email string,
	code string,
	purpose OTPPurpose,
) (*OTP, error) {

	query := `
		SELECT
			id,
			email,
			code_hash,
			purpose,
			expires_at,
			used,
			attempts,
			max_attempts,
			created_at
		FROM otp_codes
		WHERE email = $1
		  AND purpose = $2
		  AND used = false
		  AND attempts < max_attempts
		  AND expires_at > now()
		ORDER BY created_at DESC
		LIMIT 1;
	`

	var otp OTP

	err := p.DB.QueryRow(query, email, purpose).Scan(
		&otp.ID,
		&otp.Email,
		&otp.CodeHash,
		&otp.Purpose,
		&otp.ExpiresAt,
		&otp.Used,
		&otp.Attempts,
		&otp.MaxAttempts,
		&otp.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid or expired otp")
		}
		return nil, err
	}

	if err := utils.VerifyHash(otp.CodeHash, code); err != nil {
		_, _ = p.DB.Exec(
			`UPDATE otp_codes SET attempts = attempts + 1 WHERE id = $1`,
			otp.ID,
		)
		return nil, errors.New("invalid otp")
	}

	_, err = p.DB.Exec(
		`UPDATE otp_codes SET used = true WHERE id = $1`,
		otp.ID,
	)
	if err != nil {
		return nil, err
	}

	return &otp, nil
}



