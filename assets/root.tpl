<div class="container" ng-controller="VoteCtrl">
  <a href="http://sg.uchicago.edu">
    <img src="/sg.png" style="width:3in" class="masthead hidden-xs hidden-sm"/>
  </a>
  <div class="panel panel-default">
    <div class="panel-heading" ng-show="isAdmin">
      <h3 class="panel-title">
      <a class="btn btn-default"
         href="{{ baseURL + '/results.csv' }}">Download CSV</a>
      <a class="btn btn-default vote-edit-button"
         ng-click="setEditMode(!editMode);">
           {{ editMode ? "Save" : "Edit" }}</a>
      </h3>
    </div>
    <table class="table">
      <tr ng-show="editMode">
        <td>
          <div class="btn-group" dropdown>
            <button class="btn btn-default dropdown-toggle" dropdown-toggle>
              Add Plugin...
              <span class="caret"></span></button>
            <ul class="dropdown-menu">
              <li class="{{ voteLimits ? 'disabled' : '' }}">
                <a ng-click="setOption('limits', true)">Vote Limits</a></li>
              <li class="{{ voteWeight > 0 ? 'disabled' : '' }}">
                <a ng-click="setOption('weight', true)">
                   Dollar Amount Per Vote</a></li>
              <li class="{{ voteRandomized ? 'disabled' : '' }}">
                <a ng-click="setOption('randomized', true)">
                  Randomized Voting</a></li>
              <li class="{{ showProgress ? 'disabled' : '' }}">
                <a ng-click="setOption('showProgress', true)">
                  Progress Bar</a></li>
            </ul>
          </div>
        </td>
      </tr>
      <tr ng-show="editMode && voteLimits" class="vote-plugin">
        <td>
          <label>
            Users cannot select above the limit.
          </label>
        </td>
        <td>
          <a class="btn btn-danger" ng-click="setOption('limits', false)">
            Disable
          </a>
        </td>
      </tr>
      <tr ng-show="editMode && voteWeight != 0" class="form-inline vote-plugin">
        <td>
          <label>
            Each selection contributes:
          </label>
        </td>
        <td>
          <input type="number" class="form-control" style="width:5em"
            ng-model="voteWeight"/>
        </td>
      </tr>
      <tr ng-show="editMode && voteRandomized" class="form-inline vote-plugin">
        <td>
          <label>
            Options are randomized before display.
          </label>
        </td>
        <td>
          <a class="btn btn-danger" ng-click="setOption('randomized', false)">
            Disable
          </a>
        </td>
      </tr>
      <tr ng-show="editMode && showProgress" class="form-inline vote-plugin">
        <td>
          <label>
            Voters can see the number of spaces left.
          </label>
        </td>
        <td>
          <a class="btn btn-danger" ng-click="setOption('showProgress', false)">
            Disable
          </a>
        </td>
      </tr>
      <tr ng-repeat="candidate in candidates |
              orderBy:(!editMode && voteRandomized ? 'order' : '')">
        <td ng-hide="editMode">
          <p><strong>{{ candidate.name }}</strong>
            {{ candidate.description }}</p>
        </td>
        <td ng-show="editMode">
          <input type="text" class="form-control"
                 ng-model="candidate.name"
                 placeholder="(name)"/>
          <textarea class="form-control"
                    ng-model="candidate.description"
                    placeholder="(description)"></textarea>
        </td>
        <td style="text-align:right" ng-hide="editMode" class="vote-right">
          <a class="btn vote-button
                      {{ vote == candidate.id ?
                          (pending ? 'btn-primary disabled' :
                            'btn-success active')
                            : ((candidate.progress + 1) * (voteWeight || 1)
                                > candidate.request ?
                                'btn-primary disabled'
                                : 'btn-primary')
                        }}"
               ng-click="voteFor(candidate)">
              {{ getVerbFor(candidate) }}
          </a>
          <div class="progress vote-progress" ng-show="candidate.request && showProgress">
            <div class="progress-bar progress-bar-success progress-bar-striped"
                 style="
                   width:{{ candidate.progress * (voteWeight || 1) * 100
                       / candidate.request }}%;">
            </div>
          </div>
          <div class="vote-caption" ng-show="candidate.request && showProgress">
            {{ candidate.progress * (voteWeight || 1) | number:0 }} of
            {{ candidate.request | number:0 }}
          </div>
        </td>
        <td ng-show="editMode" class="vote-right">
          <input type="number" class="form-control"
                 ng-model="candidate.request" placeholder="(no target)"/>
        </td>
      </tr>
      <tr ng-show="editMode">
        <td colspan="2">
            <a class="btn btn-primary"
              ng-click="addNewRow()">Add New Row</a></td>
      </tr>
    </table>
  </div>
</div>
