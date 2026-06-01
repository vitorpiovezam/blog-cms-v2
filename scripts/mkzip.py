#!/usr/bin/env python3
"""
Package the Go Lambda binary into a deployment zip.

Usage: python3 scripts/mkzip.py
  - Reads binary from  .bin/bootstrap
  - Reads posts from   src/posts/
  - Writes zip to      .bin/bootstrap.zip

Lambda expects a `bootstrap` executable at the zip root (provided.al2023).
src/posts/ is bundled so the function can read posts from disk on cold starts.
"""

import os
import stat
import zipfile

BIN = ".bin/bootstrap"
OUT = ".bin/bootstrap.zip"
POSTS = "src/posts"


def main():
    if not os.path.isfile(BIN):
        print(f"error: {BIN} not found — run `make build` first")
        raise SystemExit(1)

    with zipfile.ZipFile(OUT, "w", zipfile.ZIP_DEFLATED) as zf:
        # bootstrap binary — must be executable inside the zip
        info = zipfile.ZipInfo("bootstrap")
        info.external_attr = (
            stat.S_IRWXU | stat.S_IRGRP | stat.S_IXGRP | stat.S_IROTH | stat.S_IXOTH
        ) << 16
        with open(BIN, "rb") as f:
            zf.writestr(info, f.read())

        # bundle markdown posts (cold-start fallback)
        if os.path.isdir(POSTS):
            for fname in sorted(os.listdir(POSTS)):
                if fname.endswith(".md"):
                    zf.write(os.path.join(POSTS, fname), os.path.join("src", "posts", fname))

        # bundle static assets used by posts
        assets = "src/assets"
        if os.path.isdir(assets):
            for root, _, files in os.walk(assets):
                for fname in files:
                    full = os.path.join(root, fname)
                    arc = full.replace(os.sep, "/")
                    zf.write(full, arc)

    print(f"✓ {OUT}")


if __name__ == "__main__":
    main()
