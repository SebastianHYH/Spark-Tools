import asyncio
import json
import aiohttp  # Import aiohttp for session management
import requests
from Crypto.Cipher import AES
from Crypto.Util.Padding import unpad
import fortnite_api
import os

def decrypt_aes(key, data):
    cipher = AES.new(key, AES.MODE_ECB)
    decrypted_data = cipher.decrypt(data)
    try:
        return unpad(decrypted_data, AES.block_size)  # Ensure correct padding removal
    except ValueError:
        return None  # If padding is incorrect, return None

def filePathGrabber(json_url, json_file_path):
    # Download the JSON file
    response = requests.get(json_url)
    response.raise_for_status()  # Raise an error for HTTP issues

    # Save JSON content to a temporary file
    with open(json_file_path, "w", encoding="utf-8") as json_file:
        json_file.write(response.text)

async def main():
    async with aiohttp.ClientSession() as session:
        json_url = "https://fortnitecontent-website-prod07.ol.epicgames.com/content/api/pages/fortnite-game/spark-tracks"
        json_file_path = "temp_spark_tracks.json"  # Temporary JSON file
        client = fortnite_api.Client(session=session)
        aes = await client.fetch_aes()  # Await the coroutine

        # Extract main AES key and dynamic keys
        # aes_keys = [aes.main_key] + [key.key for key in aes.dynamic_keys]
        aes_keys = ["29b4ac18d090166559244e15548bd4c11b98d33ad57f7b0d9bfff6ceb7cf6145"]
        filePathGrabber(json_url, json_file_path)
        with open(json_file_path, 'r') as f:
            data = json.load(f)

        song_title = "Empty Out Your Pockets"  # Input your song title
        selected_song = None
        for track_key, track_data in data.items():
            if isinstance(track_data, dict) and track_data.get("track", {}).get("tt") == song_title:
                selected_song = track_data
                break

        if not selected_song:
            print("Song not found")
        else:
            music_url = selected_song["track"]["mu"]
            print(f"Found {song_title} at {music_url}, downloading...")

            response = requests.get(music_url)
            if response.status_code == 200:
                encrypted_data = response.content

                # Try decrypting with multiple keys
                decrypted_data = None
                for key in aes_keys:
                    # Convert hex key to bytes if it's a hex string
                    if isinstance(key, str):
                        key = bytes.fromhex(key)

                    decrypted_data = decrypt_aes(key, encrypted_data)
                    if decrypted_data:
                        print(f"Decrypted with key {key.hex()[:16]}...")  # Show part of the key
                        break

                if decrypted_data:
                    output_file = f"{song_title.replace(' ', '_')}.mid"
                    with open(output_file, "wb") as f:
                        f.write(decrypted_data)
                    print(f"Downloaded {song_title} successfully")
                else:
                    print(f"Failed to decrypt {song_title} with all keys")
            else:
                print(f"Failed to download {song_title}: {response.status_code}")        
        # Delete the temporary JSON file
        os.remove(json_file_path)

# Run the asynchronous main function
asyncio.run(main())
