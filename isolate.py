import sys
import os
import subprocess
import shutil

if len(sys.argv) < 4:
    print("Usage: isolate.py input output is_video")
    sys.exit(1)

input_file = sys.argv[1]
output_file = sys.argv[2]
is_video = sys.argv[3].lower() == "true"

base_dir = "/home/lennox/Desktop/اهها/Go_Bot"
ffmpeg_path = os.path.join(base_dir, "node_modules", "ffmpeg-static", "ffmpeg")
ffprobe_path = os.path.join(base_dir, "node_modules", "ffprobe-static", "bin", "linux", "x64", "ffprobe")

env = os.environ.copy()
env["PATH"] = os.path.dirname(ffmpeg_path) + os.pathsep + os.path.dirname(ffprobe_path) + os.pathsep + env.get("PATH", "")
env["FFMPEG_BINARY"] = ffmpeg_path
env["FFPROBE_BINARY"] = ffprobe_path

audio_file = "temp_audio.wav"

subprocess.run([ffmpeg_path, "-y", "-i", input_file, "-vn", "-acodec", "pcm_s16le", "-ar", "44100", audio_file], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

subprocess.run(["python3", "-m", "spleeter", "separate", "-p", "spleeter:2stems", "-o", "output", audio_file], check=True, env=env)

vocals_file = "output/temp_audio/vocals.wav"

if is_video:
    subprocess.run([ffmpeg_path, "-y", "-i", input_file, "-i", vocals_file, "-c:v", "copy", "-map", "0:v:0", "-map", "1:a:0", output_file], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
else:
    subprocess.run([ffmpeg_path, "-y", "-i", vocals_file, output_file], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

shutil.rmtree("output", ignore_errors=True)
if os.path.exists(audio_file):
    os.remove(audio_file)
