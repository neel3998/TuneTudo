// static/js/app.js

let currentAudio = null;
let currentSongId = null;

// Modal functions
function showLoginModal() {
    document.getElementById('loginModal').classList.add('show');
}

function closeLoginModal() {
    document.getElementById('loginModal').classList.remove('show');
}

function showRegisterModal() {
    closeLoginModal();
    document.getElementById('registerModal').classList.add('show');
}

function closeRegisterModal() {
    document.getElementById('registerModal').classList.remove('show');
}

// Authentication functions
async function login(event) {
    event.preventDefault();
    
    const username = document.getElementById('loginUsername').value;
    const password = document.getElementById('loginPassword').value;
    const isAdminCheck = document.getElementById('isAdmin').checked;
    
    try {
        const response = await AuthAPI.login(username, password);
        
        if (isAdminCheck && !response.data.user.is_admin) {
            showAlert('No admin privileges', 'error');
            return;
        }
        
        setAuthToken(response.data.token);
        setUser(response.data.user);
        
        showAlert('Login successful!');
        closeLoginModal();
        
        setTimeout(() => {
            if (response.data.user.is_admin) {
                window.location.href = 'admin.html';
            } else {
                window.location.reload();
            }
        }, 1000);
    } catch (error) {
        showAlert(error.message, 'error');
    }
}

async function register(event) {
    event.preventDefault();
    
    const username = document.getElementById('regUsername').value;
    const email = document.getElementById('regEmail').value;
    const password = document.getElementById('regPassword').value;
    
    try {
        await AuthAPI.register(username, email, password);
        showAlert('Registration successful! Please login.');
        closeRegisterModal();
        showLoginModal();
    } catch (error) {
        showAlert(error.message, 'error');
    }
}

// Search function
async function performSearch() {
    const query = document.getElementById('searchInput').value.trim();
    
    if (!query) {
        showAlert('Please enter a search query', 'error');
        return;
    }
    
    try {
        const response = await SearchAPI.search(query);
        displaySearchResults(response.data);
    } catch (error) {
        showAlert('Search temporarily unavailable', 'error');
    }
}

function displaySearchResults(results) {
    const resultsDiv = document.getElementById('searchResults');
    
    if (!results.songs.length && !results.artists.length && !results.albums.length) {
        resultsDiv.innerHTML = '<p class="empty-state">No results found</p>';
        return;
    }
    
    let html = '<h3>Search Results</h3>';
    
    if (results.songs && results.songs.length > 0) {
        html += '<h4>Songs</h4><div class="songs-grid">';
        results.songs.forEach(song => {
            html += createSongCard(song);
        });
        html += '</div>';
    }
    
    if (results.artists && results.artists.length > 0) {
        html += '<h4>Artists</h4><div class="songs-grid">';
        results.artists.forEach(artist => {
            html += `
                <div class="song-card">
                    <h4>${artist.name}</h4>
                    <p>${artist.description || 'Artist'}</p>
                </div>
            `;
        });
        html += '</div>';
    }
    
    if (results.albums && results.albums.length > 0) {
        html += '<h4>Albums</h4><div class="songs-grid">';
        results.albums.forEach(album => {
            html += `
                <div class="song-card">
                    <h4>${album.title}</h4>
                    <p>${album.artist ? album.artist.name : 'Unknown Artist'}</p>
                </div>
            `;
        });
        html += '</div>';
    }
    
    resultsDiv.innerHTML = html;
}

function createSongCard(song) {
    const duration = formatDuration(song.duration_seconds);
    return `
        <div class="song-card">
            <h4>${song.title}</h4>
            <p>${song.artist ? song.artist.name : 'Unknown Artist'}</p>
            ${song.album ? `<p>${song.album.title}</p>` : ''}
            <p>${duration}</p>
            <div class="song-actions">
                <button class="btn btn-primary" onclick="playSong(${song.id}, '${escapeHtml(song.title)}', '${escapeHtml(song.artist ? song.artist.name : 'Unknown')}')">Play</button>
                ${isAuthenticated() ? `<button class="btn btn-secondary" onclick="showAddToPlaylistModal(${song.id})">Add to Playlist</button>` : ''}
            </div>
        </div>
    `;
}

function formatDuration(seconds) {
    if (!seconds) return '--:--';
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML.replace(/'/g, "\\'");
}

// Load categories
async function loadCategories() {
    try {
        const response = await SearchAPI.getCategories();
        displayCategories(response.data);
    } catch (error) {
        console.error('Failed to load categories:', error);
    }
}

function displayCategories(categories) {
    const grid = document.getElementById('categoriesGrid');
    
    if (!categories || categories.length === 0) {
        grid.innerHTML = '<p>No categories available</p>';
        return;
    }
    
    grid.innerHTML = categories.map(cat => `
        <div class="category-card" onclick="viewCategory(${cat.id}, '${escapeHtml(cat.name)}')">
            <h3>${cat.name}</h3>
            <p>${cat.description || ''}</p>
        </div>
    `).join('');
}

async function viewCategory(categoryId, categoryName) {
    try {
        const response = await SearchAPI.getSongsByCategory(categoryId);
        
        if (response.data.length === 0) {
            showAlert('No tracks available in this category');
            return;
        }
        
        const resultsDiv = document.getElementById('searchResults');
        resultsDiv.innerHTML = `
            <h3>${categoryName}</h3>
            <div class="songs-grid">
                ${response.data.map(song => createSongCard(song)).join('')}
            </div>
        `;
        
        resultsDiv.scrollIntoView({ behavior: 'smooth' });
    } catch (error) {
        showAlert('Failed to load category songs', 'error');
    }
}

// Load recent songs
async function loadRecentSongs() {
    try {
        const response = await SongsAPI.getRecent(12);
        displayRecentSongs(response.data);
    } catch (error) {
        console.error('Failed to load recent songs:', error);
    }
}

function displayRecentSongs(songs) {
    const grid = document.getElementById('recentSongs');
    
    if (!songs || songs.length === 0) {
        grid.innerHTML = '<p>No songs available</p>';
        return;
    }
    
    grid.innerHTML = songs.map(song => createSongCard(song)).join('');
}

// Audio player functions
function playSong(songId, title, artist) {
    const audioPlayer = document.getElementById('audioPlayer');
    const audioElement = document.getElementById('audioElement');
    const titleElement = document.getElementById('playerSongTitle');
    const artistElement = document.getElementById('playerArtist');
    
    currentSongId = songId;
    titleElement.textContent = title;
    artistElement.textContent = artist;
    
    audioElement.src = SongsAPI.getStreamUrl(songId);
    audioElement.play();
    
    audioPlayer.style.display = 'block';
    currentAudio = audioElement;
    
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
    currentAudio = null;
    currentSongId = null;
}

// Add to playlist modal
let selectedSongIdForPlaylist = null;

async function showAddToPlaylistModal(songId) {
    if (!isAuthenticated()) {
        showAlert('Please login to add songs to playlists', 'error');
        return;
    }
    
    selectedSongIdForPlaylist = songId;
    
    try {
        const response = await PlaylistAPI.getUserPlaylists();
        const playlists = response.data;
        
        if (!playlists || playlists.length === 0) {
            showAlert('No playlists found. Create a playlist first!', 'error');
            return;
        }
        
        const modal = document.createElement('div');
        modal.id = 'tempPlaylistModal';
        modal.className = 'modal show';
        modal.innerHTML = `
            <div class="modal-content">
                <span class="close" onclick="closeTempModal()">&times;</span>
                <h2>Add to Playlist</h2>
                <div class="playlist-selection">
                    ${playlists.map(pl => `
                        <div class="playlist-selection-item" onclick="addToPlaylist(${pl.id})">
                            <h4>${pl.name}</h4>
                            <p>${pl.song_count || 0} songs</p>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
    } catch (error) {
        showAlert('Failed to load playlists', 'error');
    }
}

async function addToPlaylist(playlistId) {
    try {
        await PlaylistAPI.addSongToPlaylist(playlistId, selectedSongIdForPlaylist);
        showAlert('Song added to playlist!');
        closeTempModal();
    } catch (error) {
        showAlert(error.message, 'error');
    }
}

function closeTempModal() {
    const modal = document.getElementById('tempPlaylistModal');
    if (modal) modal.remove();
}

// Search on Enter key
document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('searchInput');
    if (searchInput) {
        searchInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                performSearch();
            }
        });
    }
    
    // Audio element event listeners
    const audioElement = document.getElementById('audioElement');
    if (audioElement) {
        audioElement.addEventListener('play', updatePlayPauseButton);
        audioElement.addEventListener('pause', updatePlayPauseButton);
    }
    
    // Load initial data
    loadCategories();
    loadRecentSongs();
});