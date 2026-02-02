// static/js/playlists.js

// Check authentication
if (!isAuthenticated()) {
    window.location.href = 'index.html';
}

let currentPlaylistId = null;

// Load user playlists
async function loadPlaylists() {
    try {
        const response = await PlaylistAPI.getUserPlaylists();
        displayPlaylists(response.data);
    } catch (error) {
        showAlert('Failed to load playlists', 'error');
    }
}

function displayPlaylists(playlists) {
    const grid = document.getElementById('playlistsGrid');
    const noPlaylists = document.getElementById('noPlaylists');
    
    if (!playlists || playlists.length === 0) {
        grid.style.display = 'none';
        noPlaylists.style.display = 'block';
        return;
    }
    
    grid.style.display = 'grid';
    noPlaylists.style.display = 'none';
    
    grid.innerHTML = playlists.map(playlist => `
        <div class="playlist-card" onclick="viewPlaylistDetails(${playlist.id})">
            <h3>${playlist.name}</h3>
            <p>${playlist.description || 'No description'}</p>
            <p class="text-secondary">${playlist.song_count || 0} songs</p>
        </div>
    `).join('');
}

// Create playlist modal
function showCreatePlaylistModal() {
    document.getElementById('createPlaylistModal').classList.add('show');
}

function closeCreatePlaylistModal() {
    document.getElementById('createPlaylistModal').classList.remove('show');
    document.getElementById('createPlaylistForm').reset();
}

async function createPlaylist(event) {
    event.preventDefault();
    
    const name = document.getElementById('playlistName').value;
    const description = document.getElementById('playlistDescription').value || null;
    
    try {
        await PlaylistAPI.createPlaylist(name, description);
        showAlert('Playlist created successfully!');
        closeCreatePlaylistModal();
        loadPlaylists();
    } catch (error) {
        showAlert(error.message, 'error');
    }
}

// View playlist details
async function viewPlaylistDetails(playlistId) {
    currentPlaylistId = playlistId;
    
    try {
        const response = await PlaylistAPI.getPlaylistDetails(playlistId);
        displayPlaylistDetails(response.data);
    } catch (error) {
        showAlert('Failed to load playlist details', 'error');
    }
}

function displayPlaylistDetails(data) {
    const modal = document.getElementById('playlistDetailsModal');
    const nameElement = document.getElementById('playlistDetailsName');
    const descElement = document.getElementById('playlistDetailsDescription');
    const songsContainer = document.getElementById('playlistSongs');
    const noSongs = document.getElementById('noSongs');
    
    nameElement.textContent = data.playlist.name;
    descElement.textContent = data.playlist.description || 'No description';
    
    if (!data.songs || data.songs.length === 0) {
        songsContainer.style.display = 'none';
        noSongs.style.display = 'block';
    } else {
        songsContainer.style.display = 'block';
        noSongs.style.display = 'none';
        
        songsContainer.innerHTML = data.songs.map(item => {
            const song = item.song;
            return `
                <div class="playlist-song-item">
                    <div>
                        <h4>${song.title}</h4>
                        <p>${song.artist ? song.artist.name : 'Unknown Artist'}</p>
                    </div>
                    <div class="song-actions">
                        <button class="btn btn-primary" onclick="playSong(${song.id}, '${escapeHtml(song.title)}', '${escapeHtml(song.artist ? song.artist.name : 'Unknown')}')">Play</button>
                        <button class="btn btn-danger" onclick="removeSongFromPlaylist(${song.id})">Remove</button>
                    </div>
                </div>
            `;
        }).join('');
    }
    
    modal.classList.add('show');
}

function closePlaylistDetailsModal() {
    document.getElementById('playlistDetailsModal').classList.remove('show');
    currentPlaylistId = null;
}

// Remove song from playlist
async function removeSongFromPlaylist(songId) {
    if (!currentPlaylistId) return;
    
    if (!confirm('Remove this song from the playlist?')) return;
    
    try {
        await PlaylistAPI.removeSongFromPlaylist(currentPlaylistId, songId);
        showAlert('Song removed from playlist');
        viewPlaylistDetails(currentPlaylistId);
    } catch (error) {
        showAlert(error.message, 'error');
    }
}

// Delete playlist
async function deleteCurrentPlaylist() {
    if (!currentPlaylistId) return;
    
    if (!confirm('Are you sure you want to delete this playlist?')) return;
    
    try {
        await PlaylistAPI.deletePlaylist(currentPlaylistId);
        showAlert('Playlist deleted successfully');
        closePlaylistDetailsModal();
        loadPlaylists();
    } catch (error) {
        showAlert(error.message, 'error');
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

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML.replace(/'/g, "\\'");
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    loadPlaylists();
    
    const audioElement = document.getElementById('audioElement');
    if (audioElement) {
        audioElement.addEventListener('play', updatePlayPauseButton);
        audioElement.addEventListener('pause', updatePlayPauseButton);
    }
});