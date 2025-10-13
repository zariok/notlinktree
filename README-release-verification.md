# Verifying Releases

All official releases of NotLinkTree are signed with a dedicated GPG key. You can verify the integrity of the downloaded files to ensure they have not been tampered with.

## 1. Import the Public Key

First, you need to import the NotLinkTree release signing key into your local GPG keyring.

**Public Key for Release Signing:**
```
-----BEGIN PGP PUBLIC KEY BLOCK-----

mDMEaO0q7RYJKwYBBAHaRw8BAQdAoz4h2WiZhfoLuJh0278+pze1/4XQPfNzATkA
mB1MOpu0WURyYXVnaG4uZGV2IFNvZnR3YXJlIFNpZ25pbmcgS2V5IChTb2Z0d2Fy
ZSBTaWduaW5nIGFuZCBSZWxlYXNlIEtleSkgPHN1cHBvcnRAZHJhdWdobi5kZXY+
iJkEExYKAEEWIQSUgSyNs3/t8dwxMHXg3bHgnicJ/QUCaO0q7QIbAwUJA8JnAAUL
CQgHAgIiAgYVCgkICwIEFgIDAQIeBwIXgAAKCRDg3bHgnicJ/da6AP9dnUnt8FSY
pbE1LPiqs395EyHqzjKq3MHmELXIYZWIMwD/bM+Z2QuteuNqUvhMU1HzmomO/1hE
whR00vfLVaGvPwq4MwRo7StvFgkrBgEEAdpHDwEBB0AfgQ1Ks00pWoU2IKFoFUV/
swgkK9B5lvaCk62PcSvBg4j1BBgWCgAmFiEElIEsjbN/7fHcMTB14N2x4J4nCf0F
AmjtK28CGwIFCQPCZwAAgQkQ4N2x4J4nCf12IAQZFgoAHRYhBF5/OJmO3c7mUzPD
DzjLCw0fsI/xBQJo7StvAAoJEDjLCw0fsI/xGlYA/iwsXffYxSMHFv2vahc+ISRp
AD7heKAQvrXQUmfozTH9AQChrCK1n6eInv7oLkBWBoQfHl3lZrdokovY8ElgnWT9
DmZoAPwJWOjt54XzoFkm3p0kgR8qKog84lPXtIp6qAtV0z5lJAD/R2X4LwtamwWM
cbQANYouIUHtD/YmTzz3w1p0APIzTgM=
=yfho
-----END PGP PUBLIC KEY BLOCK-----
```

You can import this key by saving it to a file (e.g., `release-key.asc`) and running:
```bash
gpg --import release-key.asc
```

## 2. Download Release Files

From the [Releases page](https://github.com/zariok/notlinktree/releases), download the binary for your platform, along with the `checksums.txt` and `checksums.txt.asc` files.

## 3. Verify the Signature

Verify that the checksums file was signed by the key you imported:
```bash
gpg --verify checksums.txt.asc checksums.txt
```
You should see output indicating a "Good signature" from the "Draughn Signing Key". A warning about the key not being certified by a trusted signature is expected unless you have established a web of trust.

## 4. Verify the Binary

Finally, check that the checksum of your downloaded binary matches the one listed in the `checksums.txt` file.

On Linux or macOS:
```bash
sha256sum --check checksums.txt --ignore-missing
```

On Windows (using PowerShell):
```powershell
(Get-FileHash -Algorithm SHA256 (Get-Content checksums.txt -Raw).split() | ForEach-Object { $_.Hash -eq $_.Path.Split(" ")[0] })
```

If the checksums match, you have successfully verified that the binary is authentic and has not been altered.

## See Also

For general setup and configuration, see [README.md](README.md).

