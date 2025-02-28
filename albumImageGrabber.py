import json
import os
import requests

# Define the source URL for the JSON data
json_url = "https://fortnitecontent-website-prod07.ol.epicgames.com/content/api/pages/fortnite-game/spark-tracks"
json_file_path = "temp_spark_tracks.json"  # Temporary JSON file
output_folder = "album_images"  # Folder to save images

# Ensure the output folder exists
os.makedirs(output_folder, exist_ok=True)

try:
    # Download the JSON file
    response = requests.get(json_url)
    response.raise_for_status()  # Raise an error for HTTP issues

    # Save JSON content to a temporary file
    with open(json_file_path, "w", encoding="utf-8") as json_file:
        json_file.write(response.text)

    # Load JSON data from the file
    with open(json_file_path, "r", encoding="utf-8") as file:
        data = json.load(file)

    # Delete the temporary JSON file
    os.remove(json_file_path)

except requests.RequestException as e:
    print(f"Failed to download JSON file: {e}")
    exit()
except Exception as e:
    print(f"Error processing JSON file: {e}")
    exit()

# Helper function to extract album names ('tt') and album URLs ('au')
def find_album_data(obj, album_data, current_name=None):
    if isinstance(obj, dict):
        for key, value in obj.items():
            if key == "tt" and isinstance(value, str):
                current_name = value  # Update the current album name
            if key == "au" and isinstance(value, str) and current_name:
                album_data[value] = current_name  # Map URL to its corresponding name
            else:
                find_album_data(value, album_data, current_name)
    elif isinstance(obj, list):
        for item in obj:
            find_album_data(item, album_data, current_name)

# Store album URLs and names in a dictionary {url: name}
album_data = {}
find_album_data(data, album_data)

# Download images and rename them with their album names
if album_data:
    for image_url, album_name in album_data.items():
        # Replace invalid characters in file names
        safe_album_name = "".join(c for c in album_name if c.isalnum() or c in (" ", "-", "_")).rstrip()
        file_extension = os.path.splitext(image_url)[1]  # Get the image file extension
        file_name = os.path.join(output_folder, f"{safe_album_name}{file_extension}")
        image_name = os.path.join(f"{safe_album_name}{file_extension}")

        try:
            # Download and save the image
            response = requests.get(image_url)
            response.raise_for_status()
            if image_name in os.listdir(output_folder):
                print(f"{image_name} already exists.")
                continue
            else:
                with open(file_name, "wb") as img_file:
                    img_file.write(response.content)
                print(f"Downloaded: {file_name}")
        except requests.RequestException as e:
            print(f"Failed to download {image_url}: {e}")
else:
    print("No album images found in the JSON data.")
