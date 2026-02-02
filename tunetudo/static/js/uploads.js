// static/js/uploads.js

// Check authentication
if (!isAuthenticated()) {
    window.location.href = 'index.html';
}

// Load user uploads
async function loadUploads() {
    try {
        const response = await UserAPI.getUserUploads();
        displayUploads(response.data);
    } catch (error) {
        showAlert('Failed to load uploads', 'error');
    }
}

function displayUploads(songs) {
    const grid = document.getElementById('uploadsGrid');
    const noUploads = document.getElementById('noUploads');
    
    if (!songs || songs.length === 0) {
        grid.style.display = 'none';
        noUploads.style.display = 'block';
        return;
    }
    
    grid.style.display = 'grid';
    noUploads.style.display = 'none';
    
    grid.innerHTML = songs.map(song => {
        const duration = formatDuration(song.duration_seconds);
        return `
            <div class="song-card">
                <h4>${song.title}</h4>
                <p>${song.artist ? song.artist.name : 'Unknown Artist'}</p>
                <p>${duration}</p>
                <p class="text-secondary">${song.format.toUpperCase()}</p>
                <div class="song-actions">
                    <button class="btn btn-primary" onclick="playSong(${song.id}, '${(song.title)}', '${(song.artist ? song.artist.name : 'Unknown')}')">Play</button>
                </div>
            </div>
        `;
    }).join('');
}

function formatDuration(seconds) {
    if (!seconds) return '--:--';
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
}

// Upload modal
function showUploadModal() {
    document.getElementById('uploadModal').classList.add('show');
}

function closeUploadModal() {
    document.getElementById('uploadModal').classList.remove('show');
    document.getElementById('uploadForm').reset();
    document.getElementById('uploadProgress').style.display = 'none';
}

async function uploadTrack(event) {
    event.preventDefault();
    
    const fileInput = document.getElementById('audioFile');
    const file = fileInput.files[0];
    
    if (!file) {
        showAlert('Please select a file', 'error');
        return;
    }
    
    // Validate file size (50MB)
    if (file.size > 50 * 1024 * 1024) {
        showAlert('File too large. Maximum size is 50MB', 'error');
        return;
    }
    
    // Validate file type
    const validTypes = ['.mp4', '.wav', '.mp3'];
    const fileName = file.name.toLowerCase();
    const isValid = validTypes.some(type => fileName.endsWith(type));
    
    if (!isValid) {
        showAlert('Unsupported file format. Only MP4, WAV, and MP3 allowed', 'error');
        return;
    }
    
    const uploadBtn = document.getElementById('uploadBtn');
    const progressDiv = document.getElementById('uploadProgress');
    const progressFill = document.getElementById('progressBarFill');
    const progressText = document.getElementById('progressText');
    
    uploadBtn.disabled = true;
    progressDiv.style.display = 'block';
    progressFill.style.width = '0%';
    progressText.textContent = 'Uploading...';
    
    try {
        // Simulate progress (since we can't track actual upload progress with fetch)
        let progress = 0;
        const interval = setInterval(() => {
            progress += 10;
            if (progress <= 90) {
                progressFill.style.width = progress + '%';
            }
        }, 200);
        
        await UserAPI.uploadSong(file);
        
        clearInterval(interval);
        progressFill.style.width = '100%';
        progressText.textContent = 'Upload complete!';
        
        showAlert('Track uploaded successfully!');
        closeUploadModal();
        loadUploads();
    } catch (error) {
        showAlert(error.message, 'error');
    } finally {
        uploadBtn.disabled = false;
        setTimeout(() => {
            progressDiv.style.display = 'none';
        }, 2000);
    }
}

// Audio player functions
function playSong(songId, title, artist) {
    const audioPlayer = document.getElementById('audioPlayer');
    const audioElement = document.getElementById('audioElement');
    const titleElement = document.getElementById('playerSongTitle');
    const artistElement = document.getElementById('playerArtist');
    
    titleElement.textContent = title;
    artistElement.textContent = artist;
    
    audioElement.src = SongsAPI.getStreamUrl(songId);
    audioElement.play();
    
    audioPlayer.style.display = 'block';
    
    updatePlayPauseButton();
}

function playPause() {
    const audioElement = document.getElementById('audioElement');
    
    if (audioElement.paused) {
        audioElement.play();
    } else {
        audioElement.pause();
    }
    
    updatePlayPauseButton();
}

function updatePlayPauseButton() {
    const btn = document.getElementById('playPauseBtn');
    const audioElement = document.getElementById('audioElement');
    btn.textContent = audioElement.paused ? '▶️' : '⏸️';
}

function closePlayer() {
    const audioPlayer = document.getElementById('audioPlayer');
    const audioElement = document.getElementById('audioElement');
    
    audioElement.pause();
    audioElement.src = '';
    audioPlayer.style.display = 'none';
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    loadUploads();
    
    const audioElement = document.getElementById('audioElement');
    if (audioElement) {
        audioElement.addEventListener('play', updatePlayPauseButton);
        audioElement.addEventListener('pause', updatePlayPauseButton);
    }
});