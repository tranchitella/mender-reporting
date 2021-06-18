#!/usr/bin/env python

import argparse
import sys
from elasticsearch import Elasticsearch

import index
import indexer

def cmd_index(es, num, iname, schema):
    indexer.index(es, num, iname, schema)

def cmd_create(es, iname, schema):
    if schema == "flat":
        es.indices.create(index = iname, body = index.mapping_flat)
    if schema == "flat_typed":
        es.indices.create(index = iname, body = index.mapping_flat_typed)

def cmd_delete(es, iname):
    es.indices.delete(index = iname)

def main():
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "--addr",
        help="elastisearch address (default http://localhost:9200)",
        default="http://localhost:9200",
    )
    parser.add_argument(
        "--index",
        help="elastisearch index"
    )
    parser.add_argument(
        "--schema",
        help="index schema (flat|flat_typed)"
    )
    subparsers = parser.add_subparsers(dest="command")

    parser_index = subparsers.add_parser(
        "index",
        help="index <num> random devices",
    )
    parser_index.add_argument(
        "--num",
        type=int,
        default=100,
        help="number of devices (default 100)"
    )

    parser_create = subparsers.add_parser(
        "create",
        help="create index",
    )

    parser_create = subparsers.add_parser(
        "delete",
        help="delete index",
    )

    args = parser.parse_args()
    if not args.command:
        parser.parse_args(["--help"])
        sys.exit(0)

    es = Elasticsearch([{'host': 'localhost', 'port': 9200}])

    if args.command == "index":
        cmd_index(es, args.num, args.index, args.schema)

    if args.command == "create":
        cmd_create(es, args.index, args.schema)

    if args.command == "delete":
        cmd_delete(es, args.index)

if __name__ == "__main__":
    main()
