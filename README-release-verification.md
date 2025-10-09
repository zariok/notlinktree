# Verifying Releases

All official releases of NotLinkTree are signed with a dedicated GPG key. You can verify the integrity of the downloaded files to ensure they have not been tampered with.

## 1. Import the Public Key

First, you need to import the NotLinkTree release signing key into your local GPG keyring.

**Public Key for Release Signing:**
```
-----BEGIN PGP PUBLIC KEY BLOCK-----

mDMEaFitChYJKwYBBAHaRw8BAQdAfroT4wZOF4kJ1qTX6PFRgBR6HFgsgEdcwKRw
8WuArmm0KURyYXVnaG4gU2lnbmluZyBLZXkgPHN1cHBvcnRAZHJhdWdobi5kZXY+
iJMEExYKADsWIQT8Dmh466XMYKyGkwbw21vNOOH8dwUCaFitCgIbAwULCQgHAgIi
AgYVCgkICwIEFgIDAQIeBwIXgAAKCRDw21vNOOH8d2uoAQCv154eKcCYRi/65wJe
rdPREOaYDkNsyMxpAo/0BiVa6AD/dFgZUno4clkIUS5g+bvb4Sgwl1CYYnFxQwkP
v1xpDw+4OARoWK0KEgorBgEEAZdVAQUBAQdAXd04mI4FZsIu71opHF5GEUknhmAM
vDP12QZcgXaVR1YDAQgHiHgEGBYKACAWIQT8Dmh466XMYKyGkwbw21vNOOH8dwUC
aFitCgIbDAAKCRDw21vNOOH8d/wbAQDeckvAis2z/LxjQzpPus3h+qiy8+hdMil6
IODqIZP8GwD/T0/DNLuTV5EqM05hVI3vT30UpvyUCGJvtagUjf9iTAi4MwRoWK42
FgkrBgEEAdpHDwEBB0ACh5jyvuUMpSElLarw5KGVoEyAu/Jc6BmRsfD6n6dKoojv
BBgWCgAgFiEE/A5oeOulzGCshpMG8NtbzTjh/HcFAmhYrjYCGwIAgQkQ8NtbzTjh
/Hd2IAQZFgoAHRYhBLskcU2EbvDCrgkVDXk3Sdwk1l6xBQJoWK42AAoJEHk3Sdwk
1l6xcGgA/iShdzoDjhfPIj9cxWPhPcAzWyfxtYvbMWZkegeUDA8+AP9XLHEUXYCj
ggUiBqerlvk9DNZQGI2UiJ/bUE78l6qOCFxiAP0d2unUDrVNKpUIu7lWd36KbFE2
5RhlqaZ5s1k2QO/EqgEAtRvEqaMtk5pj45RplnaPGHd6XIPNlPfRRI4cg7aVuw8=
=Xop0
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

