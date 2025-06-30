1. Run `.venv\Scripts\activate`
2. Run `pip install -r src/requirements.txt`
3. Run `uvicorn src.main:app --reload` (Optional param `--port 8000`)

## For debugging
Vscode has FastAPI support, just look for the preset `Python Debugger: FastAPI` in the debug tab.