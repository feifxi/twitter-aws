import os
import json
import psycopg2
import google.generativeai as genai

# Setup Google Gemini
GENAI_API_KEY = os.environ.get("GEMINI_API_KEY")
genai.configure(api_key=GENAI_API_KEY)

# Setup Postgres — reuse connection across warm Lambda invocations
DB_DSN = os.environ.get("DATABASE_URL")
if not DB_DSN:
    raise ValueError("DATABASE_URL environment variable is required")

_connection = None

def get_db_connection():
    """Return a reusable DB connection, reconnecting if stale."""
    global _connection
    if _connection is not None:
        try:
            _connection.cursor().execute("SELECT 1")
            return _connection
        except Exception:
            try:
                _connection.close()
            except Exception:
                pass
            _connection = None
    _connection = psycopg2.connect(DB_DSN)
    return _connection

def process_message(conn, msg_body):
    data = json.loads(msg_body)
    tweet_id = data.get("tweet_id")
    content = data.get("content")

    if not tweet_id or not content:
        print(f"Skipping invalid message: {msg_body}")
        return

    print(f"Processing embedding for tweet {tweet_id}")

    # Call Gemini API
    response = genai.embed_content(
        model="models/text-embedding-004",
        content=content,
        task_type="retrieval_document"
    )
    
    embedding_val = response.get('embedding')
    if not embedding_val:
        raise ValueError(f"Failed to extract embedding vector from response: {response}")
        
    # Convert embedding to a readable format for psycopg2 (list formatting)
    VECTOR_STR = f"[{','.join(map(str, embedding_val))}]"
        
    with conn.cursor() as cur:
         cur.execute(
             """
             INSERT INTO tweet_embeddings (tweet_id, content, embedding)
             VALUES (%s, %s, %s)
             ON CONFLICT DO NOTHING
             """,
             (tweet_id, content, VECTOR_STR)
         )
    conn.commit()
    print(f"Successfully saved embedding for tweet {tweet_id}")

def lambda_handler(event, context):
    conn = get_db_connection()
    batch_item_failures = []

    records = event.get('Records', [])
    for record in records:
        try:
            process_message(conn, record['body'])
        except Exception as e:
            if "429" in str(e) or "RESOURCE_EXHAUSTED" in str(e).upper():
                print(f"Rate limited by Gemini for message {record['messageId']}: {e}")
            else:
                print(f"Error processing message {record['messageId']}: {e}")
                conn.rollback()
            batch_item_failures.append({
                "itemIdentifier": record["messageId"]
            })

    return {
        "batchItemFailures": batch_item_failures
    }
