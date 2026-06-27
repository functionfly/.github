package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func populateUserCoreIdent(user *User, username, companyName, role sql.NullString) {
	if username.Valid {
		user.Username = &username.String
	}
	if companyName.Valid {
		user.CompanyName = &companyName.String
	}
	if role.Valid {
		user.Role = role.String
	}
}

func populateUserVerification(user *User, verificationToken sql.NullString, verificationExpiresAt sql.NullTime) {
	if verificationToken.Valid {
		user.VerificationToken = &verificationToken.String
	}
	if verificationExpiresAt.Valid {
		user.VerificationExpiresAt = &verificationExpiresAt.Time
	}
}

func populateUserOAuth(user *User, provider, providerID sql.NullString, providerData []byte) error {
	if provider.Valid {
		user.Provider = &provider.String
	}
	if providerID.Valid {
		user.ProviderID = &providerID.String
	}
	if len(providerData) > 0 {
		if err := json.Unmarshal(providerData, &user.ProviderData); err != nil {
			return fmt.Errorf("failed to parse provider data: %w", err)
		}
	}
	return nil
}

func populateUserMFA(user *User, mfaSecret sql.NullString, mfaBackupCodes []byte, mfaLastUsed sql.NullTime) error {
	if mfaSecret.Valid {
		user.MFASecret = &mfaSecret.String
	}
	if len(mfaBackupCodes) > 0 {
		if err := json.Unmarshal(mfaBackupCodes, &user.MFABackupCodes); err != nil {
			return fmt.Errorf("failed to parse MFA backup codes: %w", err)
		}
	}
	if mfaLastUsed.Valid {
		user.MFALastUsed = &mfaLastUsed.Time
	}
	return nil
}

func populateUserNameBioLastActiveProfile(user *User, nameNull, bioNull sql.NullString, lastActiveNull sql.NullTime, profileNumberNull sql.NullInt64) {
	if nameNull.Valid {
		user.Name = nameNull.String
	}
	if bioNull.Valid {
		user.Bio = &bioNull.String
	}
	if lastActiveNull.Valid {
		user.LastActiveAt = &lastActiveNull.Time
	}
	if profileNumberNull.Valid {
		pn := int(profileNumberNull.Int64)
		user.ProfileNumber = &pn
	}
}

func populateUserPublicProfileFields(user *User,
	locationNull, websiteNull, jobTitleNull sql.NullString,
	twitterURLNull, githubURLNull, linkedinURLNull sql.NullString,
) {
	if locationNull.Valid {
		user.Location = &locationNull.String
	}
	if websiteNull.Valid {
		user.Website = &websiteNull.String
	}
	if jobTitleNull.Valid {
		user.JobTitle = &jobTitleNull.String
	}
	if twitterURLNull.Valid {
		user.TwitterURL = &twitterURLNull.String
	}
	if githubURLNull.Valid {
		user.GithubURL = &githubURLNull.String
	}
	if linkedinURLNull.Valid {
		user.LinkedInURL = &linkedinURLNull.String
	}
}

func populateUserFounderFields(user *User, isFounderNull sql.NullBool, founderNumberNull sql.NullInt64) {
	if isFounderNull.Valid {
		user.IsFounder = isFounderNull.Bool
	}
	if founderNumberNull.Valid {
		fn := int(founderNumberNull.Int64)
		user.FounderNumber = &fn
	}
}
