#!/usr/bin/env bash
curl -X POST -H "Content-Type: application/json" -d '[{
    "service_uuid":"0d7eae6e-aa81-4fcc-aeea-09f83a46e3d8",
    "depend_on":"1bd77a5b-e89a-4361-8ba1-0254bcb012f0"
},{
    "service_uuid":"0d7eae6e-aa81-4fcc-aeea-09f83a46e3d8",
    "depend_on":"519d7463-d4b4-4d11-b54d-8f66786970d9"
}]' "http://127.0.0.1:9980/service_center/v2/dependency/0d7eae6e-aa81-4fcc-aeea-09f83a46e3d8"