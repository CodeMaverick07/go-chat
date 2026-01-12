package email

import "fmt"

func OTPVerificationTemplate(username, otp string) (subject, htmlBody, textBody string) {
	subject = "Verify your email"

	htmlBody = fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif;">
			<h2>Hello %s 👋</h2>
			<p>Your verification code is:</p>
			<h1 style="letter-spacing: 4px;">%s</h1>
			<p>This code expires in <b>5 minutes</b>.</p>
			<p>If you didn’t request this, please ignore this email.</p>
		</div>
	`, username, otp)

	textBody = fmt.Sprintf(
		"Hello %s,\n\nYour verification code is: %s\nThis code expires in 5 minutes.\n",
		username,
		otp,
	)

	return
}



