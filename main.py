#!/usr/bin/env python3
import argparse
from pathlib import Path
import subprocess
import sys

from lex import Lexer

def main():

    parser = argparse.ArgumentParser()
    parser.add_argument("filename")

    stage_group = parser.add_mutually_exclusive_group()
    stage_group.add_argument("--lex", action="store_true")

    args = parser.parse_args()
    path = Path(args.filename)
    if not path.is_file():
        print("ERROR: input must exist and be a file")
        sys.exit(1)

    # Preprocess
    cmd = subprocess.run(["gcc", "-E", "-P", path, "-o", path.stem + ".i"])
    if not cmd.returncode == 0:
        print(cmd.stderr)
        Path(path.stem + ".i").unlink()
        sys.exit(1)

    f = open(path.stem + ".i")
    source = f.read()
    f.close()
    lexer = Lexer(source)
    try:
        tokens = lexer.scanTokens()
    except:
        print("ERROR: parsing")
        Path(path.stem + ".i").unlink()
        sys.exit(1)
    
    # Clean Up
    Path(path.stem + ".i").unlink()

if __name__ == "__main__":
    main()
