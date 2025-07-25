from fastapi import FastAPI, HTTPException  
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse
import os
from pathlib import Path

app = FastAPI(
    title="ExGB"
)
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

static_dir = Path("static")
static_dir.mkdir(exist_ok=True)

app.mount("/static",StaticFiles(directory="static"),name="static")

@app.get("/")
async def index():
    indexPath = static_dir/ "index.html"
    return FileResponse(indexPath)