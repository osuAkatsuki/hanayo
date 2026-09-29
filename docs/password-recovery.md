# Password recovery

Self-service password changes are allowed for ordinary player privileges: public,
normal login, donor, pending verification, and premium. Any other privilege
requires manual recovery through a database administrator. Aika (user ID 999)
always requires manual recovery, regardless of its current privileges.

Protected accounts receive the usual recovery email and password-entry form.
Submitting a valid token and acceptable password returns the usual success
message and login redirect, but revokes the token without updating the account's
password. Permission checks use the current database row under a transaction
lock. Other accounts continue through the normal password update.

Protected-account events are recorded with these stages:

- `request`: a reset record was committed after the email provider accepted the
  message. Acceptance does not establish email delivery.
- `view`: an unused, unexpired token matched a reset record and its form was opened.
- `redeem`: the token was revoked and the password write was blocked. The event
  includes `outcome=password_write_blocked`.

Token-use events include the reset ID, user ID, client IP, and user agent. They do
not contain the token or submitted password. Hanayo's request logger omits the
query string on `/pwreset/continue`; upstream access logging is configured
separately.

A matched token establishes possession, not how it was obtained or who submitted
it. Owners and automated email scanners can also open recovery links. Investigate
the events before attributing them to an attacker.

Staff membership is determined from current privileges; Aika is the only fixed
account exception. Manual recovery should verify identity through an established
channel and revoke existing authentication tokens and sessions.
