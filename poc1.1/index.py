mapping_flat = {
    "settings": {"number_of_shards": 1, "number_of_replicas": 1},
    "mappings": {
        "dynamic": True,
        "_source": {"enabled": True},
        "properties": {
            "id": {"type": "keyword"},
            "tenantID": {"type": "keyword"},
            "name": {"type": "keyword"},
            "groupName": {"type": "keyword"},
            "status": {"type": "keyword"},
            "createdAt": {"type": "date"},
            "updatedAt": {"type": "date"},
        },
        "dynamic_templates": [
            {
                "inventory_strings": {
                    "match_mapping_type": "string",
                    "match": "inventory_*",
                    "mapping": {"type": "keyword"},
                }
            },
            {
                "identity_strings": {
                    "match_mapping_type": "string",
                    "match": "identity_*",
                    "mapping": {"type": "keyword"},
                }
            },
            {
                "custom_strings": {
                    "match_mapping_type": "string",
                    "match": "custom_*",
                    "mapping": {"type": "keyword"},
                }
            },
            {
                "inventory_nums_long": {
                    "match_mapping_type": "long",
                    "match": "inventory_*",
                    "mapping": {"type": "double"},
                }
            },
            {
                "identity_nums_long": {
                    "match_mapping_type": "long",
                    "match": "identity_*",
                    "mapping": {"type": "double"},
                }
            },
            {
                "custom_nums_long": {
                    "match_mapping_type": "long",
                    "match": "custom_*",
                    "mapping": {"type": "double"},
                }
            },
            {
                "inventory_nums_float": {
                    "match_mapping_type": "double",
                    "match": "inventory_*",
                    "mapping": {"type": "double"},
                }
            },
            {
                "identity_nums_float": {
                    "match_mapping_type": "double",
                    "match": "identity_*",
                    "mapping": {"type": "double"},
                }
            },
            {
                "custom_nums_float": {
                    "match_mapping_type": "double",
                    "match": "custom_*",
                    "mapping": {"type": "double"},
                }
            },
        ],
    },
}

mapping_flat_typed = {
    "settings": {"number_of_shards": 1, "number_of_replicas": 1},
    "mappings": {
        "dynamic": True,
        "_source": {"enabled": True},
        "properties": {
            "id": {"type": "keyword"},
            "tenantID": {"type": "keyword"},
            "name": {"type": "keyword"},
            "groupName": {"type": "keyword"},
            "status": {"type": "keyword"},
            "createdAt": {"type": "date"},
            "updatedAt": {"type": "date"},
        },
        "dynamic_templates": [
            {
                "inventory_strings": {
                    "match": "inventory_*_str",
                    "mapping": {"type": "keyword"},
                }
            },
            {
                "identity_strings": {
                    "match": "identity_*_str",
                    "mapping": {"type": "keyword"},
                }
            },
            {
                "custom_strings": {
                    "match": "custom_*_str",
                    "mapping": {"type": "keyword"},
                }
            },
            {
                "inventory_nums_long": {
                    "match": "inventory_*_num",
                    "mapping": {"type": "double"},
                }
            },
            {
                "identity_nums_long": {
                    "match": "identity_*_num",
                    "mapping": {"type": "double"},
                }
            },
            {
                "custom_nums_long": {
                    "match": "custom_*_num",
                    "mapping": {"type": "double"},
                }
            },
        ],
    },
}
