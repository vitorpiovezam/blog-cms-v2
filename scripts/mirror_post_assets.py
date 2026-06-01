#!/usr/bin/env python3
"""
Mirror external post assets into src/assets and upload to S3.

Usage:
  python3 scripts/mirror_post_assets.py [--upload] [--profile default]
"""

from __future__ import annotations

import argparse
import hashlib
import os
import re
import subprocess
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
POSTS_DIR = ROOT / "src" / "posts"
ASSETS_DIR = ROOT / "src" / "assets"
IMAGES_DIR = ASSETS_DIR / "images" / "posts"
AUDIO_DIR = ASSETS_DIR / "audio"

S3_BUCKET = "vitorpiovezam.com"
S3_PREFIX = "assets"
S3_REGION = "us-east-2"
PUBLIC_BASE = f"https://s3.{S3_REGION}.amazonaws.com/{S3_BUCKET}/{S3_PREFIX}"

SKIP_HOSTS = {
    "codepen.io",
    "strava.com",
    "strava-embeds.com",
    "youtube.com",
    "youtu.be",
    "www.youtube.com",
    "twitter.com",
    "www.twitter.com",
    "tripadvisor.com.br",
    "www.tripadvisor.com.br",
    "dog.ceo",
    "images.dog.ceo",
    "graphql.org",
    "console.graph.cool",
    "github.com",
    "nodejs.org",
    "www.npmjs.com",
    "es6-features.org",
    "www.apollographql.com",
    "rxjs-dev.firebaseapp.com",
    "w3schools.com",
    "www.w3schools.com",
    "angular.dev",
    "open.spotify.com",
    "bandcamp.com",
    "tshamusic.bandcamp.com",
}

IMG_PATTERNS = [
    re.compile(r'!\[[^\]]*\]\(([^)]+)\)'),
    re.compile(r'<img[^>]+src=["\']([^"\']+)["\']', re.I),
]

AUDIO_PATTERN = re.compile(
    r'<div[^>]+class=["\']blog-audio-player["\'][^>]+data-src=["\']([^"\']+)["\']',
    re.I,
)


def should_skip(url: str) -> bool:
    parsed = urllib.parse.urlparse(url)
    if not parsed.scheme.startswith("http"):
        return True
    host = (parsed.netloc or "").lower()
    if "vitorpiovezam.com" in host or host.endswith(".amazonaws.com") and S3_BUCKET in url:
        return True
    for skip in SKIP_HOSTS:
        if host == skip or host.endswith("." + skip):
            return True
    return False


def filename_for(url: str, content_type: str | None = None) -> str:
    digest = hashlib.sha1(url.encode("utf-8")).hexdigest()[:12]
    path = urllib.parse.urlparse(url).path
    ext = os.path.splitext(path)[1].lower()
    if not ext or len(ext) > 6:
        if content_type and "svg" in content_type:
            ext = ".svg"
        elif content_type and "png" in content_type:
            ext = ".png"
        elif content_type and "jpeg" in content_type or content_type and "jpg" in content_type:
            ext = ".jpg"
        elif content_type and "gif" in content_type:
            ext = ".gif"
        else:
            ext = ".bin"
    base = os.path.basename(path) or "asset"
    base = re.sub(r"[^a-zA-Z0-9._-]+", "-", base).strip("-")[:60] or "asset"
    return f"{digest}-{base}{ext if ext.startswith('.') else '.' + ext}"


def download(url: str, dest: Path) -> None:
    req = urllib.request.Request(url, headers={"User-Agent": "blog-cms-v2-asset-mirror/1.0"})
    with urllib.request.urlopen(req, timeout=60) as resp:
        content_type = resp.headers.get("Content-Type", "")
        data = resp.read()
    dest.parent.mkdir(parents=True, exist_ok=True)
    dest.write_bytes(data)
    return content_type


def collect_urls() -> dict[str, str]:
    mapping: dict[str, str] = {}
    for md in sorted(POSTS_DIR.glob("*.md")):
        text = md.read_text(encoding="utf-8")
        urls = set()
        for pat in IMG_PATTERNS:
            urls.update(pat.findall(text))
        for match in AUDIO_PATTERN.finditer(text):
            urls.add(match.group(1))

        for url in sorted(urls):
            url = url.strip()
            if should_skip(url):
                continue
            if url in mapping:
                continue

            is_audio = url.lower().endswith((".mp3", ".wav", ".ogg", ".m4a"))
            target_dir = AUDIO_DIR if is_audio else IMAGES_DIR
            fname = filename_for(url)
            local = target_dir / fname
            public = f"{PUBLIC_BASE}/{'audio' if is_audio else 'images/posts'}/{fname}"

            if not local.exists():
                try:
                    print(f"downloading {url}")
                    download(url, local)
                except urllib.error.HTTPError as exc:
                    print(f"skip {url}: HTTP {exc.code}", file=sys.stderr)
                    continue
                except Exception as exc:  # noqa: BLE001
                    print(f"skip {url}: {exc}", file=sys.stderr)
                    continue
            else:
                print(f"cached {local.name}")

            mapping[url] = public
    return mapping


def rewrite_posts(mapping: dict[str, str]) -> int:
    changed = 0
    for md in sorted(POSTS_DIR.glob("*.md")):
        original = md.read_text(encoding="utf-8")
        updated = original
        for old, new in sorted(mapping.items(), key=lambda item: len(item[0]), reverse=True):
            updated = updated.replace(old, new)
        if updated != original:
            md.write_text(updated, encoding="utf-8")
            changed += 1
            print(f"updated {md.name}")
    return changed


def upload_assets(profile: str) -> None:
    cmd = [
        "aws", "s3", "sync", str(ASSETS_DIR), f"s3://{S3_BUCKET}/{S3_PREFIX}",
        "--region", S3_REGION,
        "--profile", profile,
    ]
    print("running:", " ".join(cmd))
    subprocess.run(cmd, check=True)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--upload", action="store_true")
    parser.add_argument("--profile", default="default")
    args = parser.parse_args()

    mapping = collect_urls()
    print(f"mirrored {len(mapping)} assets")
    rewrite_posts(mapping)

    if args.upload and mapping:
        upload_assets(args.profile)


if __name__ == "__main__":
    main()
