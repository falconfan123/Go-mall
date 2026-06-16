package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSessionIDSigningAndVerification(t *testing.T) {
	t.Parallel()

	sessionID := GenerateSessionID()
	require.NotEmpty(t, sessionID)

	signed := SignSessionID(sessionID, "secret")
	verified, err := VerifySessionID(signed, "secret")
	require.NoError(t, err)
	require.Equal(t, sessionID, verified)

	_, err = VerifySessionID("invalid", "secret")
	require.ErrorIs(t, err, ErrInvalidTokenFormat)

	_, err = VerifySessionID(signed, "wrong-secret")
	require.ErrorIs(t, err, ErrInvalidSignature)
}

func TestShortTokenAndLongTokenHelpers(t *testing.T) {
	t.Parallel()

	deviceID := GenerateDeviceID()
	require.NotEmpty(t, deviceID)

	shortToken := GenerateShortToken(42, deviceID, time.Minute, "secret")
	userID, gotDeviceID, expireTime, err := VerifyShortToken(shortToken, "secret")
	require.NoError(t, err)
	require.Equal(t, uint32(42), userID)
	require.Equal(t, deviceID, gotDeviceID)
	require.Greater(t, expireTime, time.Now().Unix())

	_, _, _, err = VerifyShortToken("bad.token", "secret")
	require.ErrorIs(t, err, ErrInvalidTokenFormat)

	_, _, _, err = VerifyShortToken(shortToken, "wrong-secret")
	require.ErrorIs(t, err, ErrInvalidSignature)

	expiredToken := GenerateShortToken(42, deviceID, -time.Minute, "secret")
	_, _, _, err = VerifyShortToken(expiredToken, "secret")
	require.ErrorIs(t, err, ErrTokenExpired)

	longToken := GenerateLongToken("session-1", "secret")
	sessionID, err := VerifyLongToken(longToken, "secret")
	require.NoError(t, err)
	require.Equal(t, "session-1", sessionID)
}
