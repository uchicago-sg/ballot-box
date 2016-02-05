var voteApp = angular.module('voteApp', ['ui.bootstrap']);

String.prototype.hashCode = function() {
  var hash = 0, i, chr, len;
  if (this.length == 0) return hash;
  for (i = 0, len = this.length; i < len; i++) {
    chr   = this.charCodeAt(i);
    hash  = ((hash << 5) - hash) + chr;
    hash |= 0; // Convert to 32bit integer
  }
  return hash;
};

window.UUID = {
  generate: function() {
    var s4 = function() {
      return Math.floor((1 + Math.random()) * 0x10000).toString(16)
               .substring(1);
    };
    return s4() + s4() + '-' + s4() + '-' + s4() + '-' +
      s4() + '-' + s4() + s4() + s4();
  }
};

voteApp.controller('VoteCtrl', function ($sce, $scope, $timeout, $interval, $http) {
  $scope.vote = "";
  $scope.pending = false;
  $scope.editMode = false;
  $scope.voteLimits = false;
  $scope.voteWeight = 0;
  $scope.voteRandomized = false;
  $scope.showProgress = false;
  $scope.randomSeed = ":" + Math.random().toString();
  $scope.baseURL = window.location.href;
  $scope.descriptionText = "";
  $scope.hasDescription = false;
  $scope.voteLimit = 1;
  
  $scope.updateCandidates = function(data) {
    var newCands = data.candidates || [];
    newCands.forEach(function (cand, index) {
      if (cand.request <= 0) delete cand.request;
      cand.order = (index + $scope.randomSeed).hashCode();
    });
    $scope.candidates = newCands;
    $scope.vote = data.myvote ? data.myvote : Array(data.secondaries + 1);
    $scope.voteRandomized = data.randomized;
    $scope.voteWeight = data.weight;
    $scope.voteLimits = data.limit;
    $scope.isAdmin = data.isadmin;
    $scope.showProgress = data.showProgress;
    $scope.hasDescription = !!data.description;
    $scope.descriptionText = data.description;
    $scope.voteLimit = (data.secondaries || 0) + 1;

    while ($scope.vote.length > $scope.voteLimit) {
      $scope.vote.pop();
    }
  }
  
  $scope.updateCandidates(_DATA);
  
  $interval(function() {
    if ($scope.editMode)
      return;
    
    $http.get(window.location.href + ".json").success(function(data) {
      if ($scope.editMode)
        return;
      
      $scope.updateCandidates(data);
    });
  }, 30 * 1000);
  
  $scope.setEditMode = function(newMode) {
    $scope.candidates = $scope.candidates.filter(function(cand) {
      return !!cand.name;
    });
    
    if ($scope.editMode)
      $http.post(window.location.href + ".json", {
        "candidates": $scope.candidates,
        "limit": $scope.voteLimits,
        "weight": $scope.voteWeight,
        "randomized": $scope.voteRandomized,
        "showProgress": $scope.showProgress,
        "description": $scope.descriptionText,
        "secondaries": $scope.voteLimit - 1
      }).success(function(data) {
        $scope.updateCandidates(data);
      });
    
    $scope.editMode = newMode;
  };
  
  $scope.voteFor = function(candidate, order) {
    var prev = $scope.vote.slice();

    if ($scope.pending)
      return;

    if (!candidate)
      $scope.vote = Array($scope.vote.length);

    var id = candidate ? candidate.id : null

    for (var i = 0; i < $scope.vote.length; i++) {
      if ($scope.vote[i] == id) {
        $scope.vote[i] = "";
      }
    }

    $scope.vote[order] = id;

    var i;
    for (i = 0; i < $scope.vote.length; i++) {
      if ($scope.vote[i] != "")
        break;
    }
    
    for (var j = 0; j < i; j++) {
      if ($scope.vote[j] == "") {
        $scope.vote.splice(0, 1);
        $scope.vote.push("");
        break;
      }
    }

    $scope.pending = true;

    $http({
      method: "POST",
      url: window.location.href + "/vote.json",
      data: {
        "candidates": $scope.vote
      }
    }).success(function(data) {
      $timeout(function() {
        $scope.updateCandidates(data);
        $scope.pending = false;
      }, 200);
    }).error(function(data) {
      $scope.vote = prev;
      $scope.pending = false;
    });
  };
  
  $scope.getVerbFor = function(candidate) {
    var conj = ["Vote", "Voting", "Voted"];
    
    if (candidate.request)
      if ($scope.voteWeight)
        conj = ["Fund $" + $scope.voteWeight,
                "Funding $" + $scope.voteWeight,
                "Funded $" + $scope.voteWeight];
      else
        conj = ["Select", "Selecting", "Selected"];
    
    if ($scope.vote == candidate.id)
      if ($scope.pending)
        return conj[1];
      else
        return conj[2];
    return conj[0];
  };

  $scope.getCandidateFor = function(vote) {
    for (var i = 0; $scope.candidates[i]; i++) {
      if ($scope.candidates[i].id == vote)
        return $scope.candidates[i];
    }
  };
  
  $scope.addNewRow = function() {
    $scope.candidates.push({
      id: UUID.generate(),
      name: "",
      description: "",
      progress: 0
    });
  };
  
  $scope.setOption = function(option, enabled) {
    if (option == "limits")
      $scope.voteLimits = enabled;
    else if (option == "weight")
      $scope.voteWeight = 10;
    else if (option == "randomized")
      $scope.voteRandomized = enabled;
    else if (option == "showProgress")
      $scope.showProgress = enabled;
    else if (option == "hasDescription")
      $scope.hasDescription = enabled;
    else if (option == "voteLimit")
      $scope.voteLimit = 2;
  }
  
  $scope.markdown = function(text) {
    try {
      return $sce.trustAsHtml(new showdown.Converter().makeHtml(text));
    } catch (e) {
      return text; // in case the library fails.
    }
  }
});
