package awscertificate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

const SecretHashAnnotation = "cloud-manager.kyma-project.io/secret-hash"

func checkSecretHash(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	cert := state.ObjAsAwsCertificate()

	// Calculate hash of Secret data
	currentHash := hashSecretData(state.certificateData)

	// Get stored hash from annotation
	storedHash := cert.Annotations[SecretHashAnnotation]

	// Compare hashes
	state.secretChanged = (currentHash != storedHash)

	if state.secretChanged {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Secret data has changed, certificate will be reimported")
	}

	return nil, ctx
}

func hashSecretData(data *CertificateData) string {
	h := sha256.New()
	h.Write(data.Certificate)
	h.Write(data.PrivateKey)
	if len(data.CertificateChain) > 0 {
		h.Write(data.CertificateChain)
	}
	return hex.EncodeToString(h.Sum(nil))
}
