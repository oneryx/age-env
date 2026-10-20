package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/oneryx/age-env/internal/crypt"
	"github.com/oneryx/age-env/internal/tty"
)

const (
	// identityEnv lists identity files, separated like $PATH.
	identityEnv = "AV_IDENTITY"
	// identityDataEnv holds identity content directly, e.g. a CI secret.
	identityDataEnv = "AV_IDENTITY_DATA"
)

const noIdentityHint = "no identity found: pass -i, set " + identityEnv + " or " + identityDataEnv + ", or create ~/.config/age/keys.txt"

func readPrompt(prompt string) ([]byte, error) { return tty.ReadSecret(prompt) }

// defaultIdentityPaths returns the identity files tried when neither -i nor
// the environment variables are set.
func defaultIdentityPaths() []string {
	var paths []string
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		paths = append(paths, filepath.Join(x, "age", "keys.txt"))
	}
	if home, _ := os.UserHomeDir(); home != "" {
		paths = append(paths, filepath.Join(home, ".config", "age", "keys.txt"))
		if runtime.GOOS == "darwin" {
			paths = append(paths, filepath.Join(home, "Library", "Application Support", "age", "keys.txt"))
		}
		paths = append(paths,
			filepath.Join(home, ".ssh", "id_ed25519"),
			filepath.Join(home, ".ssh", "id_rsa"),
		)
	}
	return paths
}

// loadIdentities loads identities from -i, else from $AV_IDENTITY_DATA and
// $AV_IDENTITY, else from the default locations. Only the defaults may be
// missing or invalid; they are then skipped silently.
func (o *options) loadIdentities() ([]crypt.Identity, error) {
	if len(o.identities) > 0 {
		return loadIdentityFiles(o.identities, true)
	}
	files, data := os.Getenv(identityEnv), os.Getenv(identityDataEnv)
	if files == "" && data == "" {
		return loadIdentityFiles(defaultIdentityPaths(), false)
	}
	var ids []crypt.Identity
	if data != "" {
		got, err := crypt.ParseIdentityData("$"+identityDataEnv, []byte(data), readPrompt)
		if err != nil {
			return nil, err
		}
		ids = append(ids, got...)
	}
	if files != "" {
		got, err := loadIdentityFiles(filepath.SplitList(files), true)
		if err != nil {
			return nil, fmt.Errorf("$%s: %v", identityEnv, err)
		}
		ids = append(ids, got...)
	}
	return ids, nil
}

func loadIdentityFiles(paths []string, strict bool) ([]crypt.Identity, error) {
	var ids []crypt.Identity
	seen := map[string]bool{}
	for _, p := range paths {
		if abs, err := filepath.Abs(p); err == nil {
			if seen[abs] {
				continue
			}
			seen[abs] = true
		}
		got, err := crypt.LoadIdentityFile(p, readPrompt)
		if err != nil {
			if strict {
				return nil, err
			}
			continue
		}
		ids = append(ids, got...)
	}
	return ids, nil
}
