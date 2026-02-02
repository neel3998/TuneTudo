// static/js/api.js
const API_BASE_URL = 'https://localhost:2701/api';

// Get auth token from localStorage
function getAuthToken() {
    return localStorage.getItem('authToken');
}

// Set auth token
function setAuthToken(token) {
    localStorage.setItem('authToken', token);
}

// Remove auth token
function removeAuthToken() {
    localStorage.removeItem('authToken');
    localStorage.removeItem('user');
}

// Get user data
function getUser() {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
}

// Set user data
function setUser(user) {
    localStorage.setItem('user', JSON.stringify(user));
}

// Make authenticated API request
async function apiRequest(url, options = {}) {
    const token = getAuthToken();
    const headers = {
        ...options.headers,
    };

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    if (!(options.body instanceof FormData)) {
        headers['Content-Type'] = 'application/json';
    }

    const response = await fetch(`${API_BASE_URL}${url}`, {
        ...options,
        headers,
    });

    const data = await response.json();

    if (!response.ok) {
        throw new Error(data.message || 'Request failed');
    }

    return data;
}

// Auth API
const AuthAPI = {
    async register(username, email, password) {
        return apiRequest('/auth/register', {
            method: 'POST',
            body: JSON.stringify({ username, email, password }),
        });
    },

    async login(username, password) {
        return apiRequest('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ username, password }),
        });
    },

    async logout() {
        return apiRequest('/auth/logout', {
            method: 'POST',
        });
    },

    async getProfile() {
        return apiRequest('/profile');
    },
};

// Search API
const SearchAPI = {
    async search(query) {
        return apiRequest(`/search?q=${encodeURIComponent(query)}`);
    },

    async getCategories() {
        return apiRequest('/categories');
    },

    async getSongsByCategory(categoryId) {
        return apiRequest(`/categories/${categoryId}/songs`);
    },
};

// Songs API
const SongsAPI = {
    async getSong(songId) {
        return apiRequest(`/songs/${songId}`);
    },

    async getRecent(limit = 20) {
        return apiRequest(`/songs/recent?limit=${limit}`);
    },

    getStreamUrl(songId) {
        return `${API_BASE_URL}/songs/${songId}/stream`;
    },
};

// Playlist API
const PlaylistAPI = {
    async getUserPlaylists() {
        return apiRequest('/playlists');
    },

    async createPlaylist(name, description) {
        return apiRequest('/playlists', {
            method: 'POST',
            body: JSON.stringify({ name, description }),
        });
    },

    async getPlaylistDetails(playlistId) {
        return apiRequest(`/playlists/${playlistId}`);
    },

    async addSongToPlaylist(playlistId, songId) {
        return apiRequest(`/playlists/${playlistId}/songs`, {
            method: 'POST',
            body: JSON.stringify({ song_id: songId }),
        });
    },

    async removeSongFromPlaylist(playlistId, songId) {
        return apiRequest(`/playlists/${playlistId}/songs/${songId}`, {
            method: 'DELETE',
        });
    },

    async deletePlaylist(playlistId) {
        return apiRequest(`/playlists/${playlistId}`, {
            method: 'DELETE',
        });
    },
};

// User API
const UserAPI = {
    async uploadProfilePicture(file) {
        const formData = new FormData();
        formData.append('file', file);

        return apiRequest('/profile/picture', {
            method: 'PUT',
            body: formData,
        });
    },

    async uploadSong(file) {
        const formData = new FormData();
        formData.append('file', file);

        return apiRequest('/upload', {
            method: 'POST',
            body: formData,
        });
    },

    async getUserUploads() {
        return apiRequest('/uploads');
    },
};

// Admin API
const AdminAPI = {
    async uploadSong(file, title, artist, album, categoryId, duration) {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('title', title);
        formData.append('artist', artist);
        if (album) formData.append('album', album);
        if (categoryId) formData.append('category_id', categoryId);
        if (duration) formData.append('duration', duration);

        return apiRequest('/admin/songs', {
            method: 'POST',
            body: formData,
        });
    },

    async deleteSong(songId) {
        return apiRequest(`/admin/songs/${songId}`, {
            method: 'DELETE',
        });
    },

    async getAllSongs(limit = 50, offset = 0) {
        return apiRequest(`/admin/songs?limit=${limit}&offset=${offset}`);
    },

    async getAllUsers() {
        return apiRequest('/admin/users');
    },
};

// Check if user is authenticated
function isAuthenticated() {
    return !!getAuthToken();
}

// Check if user is admin
function isAdmin() {
    const user = getUser();
    return user && user.is_admin;
}

// Show alert message
function showAlert(message, type = 'success') {
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type}`;
    alertDiv.textContent = message;
    
    const container = document.querySelector('.container');
    if (container) {
        container.insertBefore(alertDiv, container.firstChild);
        setTimeout(() => alertDiv.remove(), 5000);
    }
}

// Initialize auth state on page load
function initAuthState() {
    const user = getUser();
    const loginBtn = document.getElementById('loginBtn');
    const userMenu = document.getElementById('userMenu');
    const usernameSpan = document.getElementById('username');
    const playlistsLink = document.getElementById('playlistsLink');
    const uploadsLink = document.getElementById('uploadsLink');
    const profileLink = document.getElementById('profileLink');
    const adminLink = document.getElementById('adminLink');

    if (user && loginBtn && userMenu) {
        loginBtn.style.display = 'none';
        userMenu.style.display = 'flex';
        if (usernameSpan) usernameSpan.textContent = user.username;
        if (playlistsLink) playlistsLink.style.display = 'block';
        if (uploadsLink) uploadsLink.style.display = 'block';
        if (profileLink) profileLink.style.display = 'block';
        
        if (user.is_admin && adminLink) {
            adminLink.style.display = 'block';
        }
    }
}

// Logout function
function logout() {
    removeAuthToken();
    window.location.href = 'index.html';
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', initAuthState);